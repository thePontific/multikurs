// frontend/src/components/Chat/Chat.tsx
import React, { useState, useRef, useEffect } from 'react';
import { Button, TextField } from '@mui/material';
import { ChatMessage } from '../../types';
import { getSampleDescription } from '../../mockData';
import './Chat.css';

export const Chat: React.FC = () => {
  // Состояния
  const [step, setStep] = useState<'name' | 'sample'>('name'); // 'name' = ввод имени, 'sample' = ввод номера
  const [username, setUsername] = useState<string>('');
  const [sampleId, setSampleId] = useState<number>(1);
  const [magnification, setMagnification] = useState<number>(10);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [inputValue, setInputValue] = useState<string>('');
  const [ws, setWs] = useState<WebSocket | null>(null);

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const magnifications = [10, 40, 100, 400, 1000];

  // Автопрокрутка вниз при новых сообщениях
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  // Подключение к WebSocket-серверу
  useEffect(() => {
    // Подключаемся только если есть username и WebSocket не подключен
    if (username && !ws) {
      const socket = new WebSocket(`ws://localhost:8001?username=${encodeURIComponent(username)}`);
      
      socket.onopen = () => {
        console.log('WebSocket подключен');
        setWs(socket);
      };
      
      socket.onmessage = (event) => {
        const response = JSON.parse(event.data);
        console.log('Получено сообщение от сервера:', response);
        
        if (response.type === 'response') {
          const data = response.data;
          
          // Теперь изображения приходят с сервера (заглушки)
          const systemMessage: ChatMessage = {
            id: (Date.now() + 1).toString(),
            sender: 'Микроскоп',
            timestamp: data.timestamp,
            isOwn: false,
            error_flag: data.error_flag,
            magnification: data.magnification,
            sample_id: data.sample_id,
            images: data.payload.images,
            text: data.error_flag
              ? 'Ошибка передачи данных. Попробуйте еще раз.'
              : getSampleDescription(data.sample_id, data.magnification),
          };
          setMessages(prev => [...prev, systemMessage]);
          setIsLoading(false);
        } else if (response.type === 'error') {
          console.error('Ошибка от сервера:', response.message);
          setIsLoading(false);
        } else if (response.type === 'accepted') {
          console.log('Запрос принят:', response);
        } else if (response.type === 'connected') {
          console.log('Приветствие:', response.message);
        }
      };
      
      socket.onclose = () => {
        console.log('WebSocket отключен');
        setWs(null);
      };
      
      socket.onerror = (error) => {
        console.error('WebSocket ошибка:', error);
        setIsLoading(false);
      };
      
      return () => {
        socket.close();
      };
    }
  }, [username]);

  // Выход из чата (очищаем всё и возвращаемся к шагу ввода имени)
  const handleLogout = () => {
    // Закрываем WebSocket соединение
    if (ws) {
      ws.close();
      setWs(null);
    }
    setStep('name');
    setUsername('');
    setSampleId(1);
    setInputValue('');
    setMessages([]);
  };

  // Отправка запроса на получение снимков (через реальный WebSocket)
  const sendRequest = async (sample: number) => {
    setIsLoading(true);

    // Добавляем сообщение-запрос пользователя в чат
    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      sender: username,
      timestamp: new Date().toISOString(),
      isOwn: true,
      error_flag: false,
      magnification: magnification,
      sample_id: sample,
      text: `Объект ${sample}, ${magnification}×`,
    };
    setMessages(prev => [...prev, userMessage]);

    // Отправляем запрос через WebSocket
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'send_request',
        sender: username,
        timestamp: new Date().toISOString(),
        magnification: magnification,
        sample_id: sample,
      }));
    } else {
      console.error('WebSocket не подключен');
      setIsLoading(false);
    }
  };

  // Обработка отправки (первый раз — имя, потом — номер образца)
  const handleSend = async () => {
    if (step === 'name') {
      // Шаг 1: сохраняем имя и переходим к вводу номера образца
      if (!inputValue.trim()) {
        alert('Введите ваше имя');
        return;
      }
      setUsername(inputValue.trim());
      setInputValue('');
      setStep('sample');
    } else {
      // Шаг 2: отправляем запрос на получение снимков
      if (!inputValue.trim()) {
        alert('Введите номер исследуемого объекта');
        return;
      }
      const sampleNum = parseInt(inputValue.trim(), 10);
      if (isNaN(sampleNum) || sampleNum < 1) {
        alert('Введите корректный номер образца (число больше 0)');
        return;
      }
      setSampleId(sampleNum);
      await sendRequest(sampleNum);
      setInputValue('');
    }
  };

  // Получаем placeholder для поля ввода
  const getPlaceholder = () => {
    return step === 'name' ? 'Введите ваше имя' : 'Введите номер исследуемого объекта';
  };

  return (
    <div className="chat-container">
      {/* Хедер с кнопкой выхода (показывается только на шаге sample) */}
      <div className="chat-header">
        {step === 'sample' && (
          <Button
            className="logout-btn"
            variant="outlined"
            onClick={handleLogout}
          >
            Выход
          </Button>
        )}
        <div className="chat-title-wrapper">
          <h1 className="chat-title">Хранилище микроснимков</h1>
          <div className="green-line"></div>
        </div>
      </div>

      {/* Область сообщений */}
      <div className="messages-area">
        <div className="messages-container">
          {messages.length === 0 && step === 'name' && (
            <div className="empty-chat">
              <span>💬</span>
              <p>Введите имя и отправьте запрос на получение снимков</p>
            </div>
          )}
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`message-wrapper ${msg.isOwn ? 'own-message' : 'system-message'}`}
            >
              {!msg.isOwn && (
                <div className="avatar">
                  <div className="avatar-placeholder">🦠</div>
                </div>
              )}
              <div className="message-bubble">
                <div className="message-sender">{msg.sender}</div>
                <div className="message-content">
                  {msg.text && <div className="message-text">{msg.text}</div>}
                  {msg.images && msg.images.length > 0 && (
                    <div className="message-images">
                      {msg.images.map((img, idx) => (
                        <img
                          key={idx}
                          src={img}
                          alt={`Снимок ${idx + 1}`}
                          className="microscope-image"
                        />
                      ))}
                    </div>
                  )}
                  {msg.error_flag && !msg.images?.length && (
                    <div className="error-icon">
                      <span>⚠️</span>
                      <span className="error-text">Ошибка передачи данных</span>
                    </div>
                  )}
                </div>
                <div className="message-time">
                  {new Date(msg.timestamp).toLocaleTimeString()}
                </div>
              </div>
            </div>
          ))}
          {isLoading && (
            <div className="loading-indicator">
              <span>📡 Загрузка снимков...</span>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      </div>

      {/* Панель ввода */}
      <div className="input-panel">
        <div className="input-row">
          <TextField
            fullWidth
            variant="outlined"
            placeholder={getPlaceholder()}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyPress={(e) => {
              if (e.key === 'Enter') {
                handleSend();
              }
            }}
            size="medium"
            sx={{
              '& .MuiOutlinedInput-root': {
                color: 'white',
                backgroundColor: '#1a1a1a',
                borderRadius: '30px',
                '& fieldset': { borderColor: '#2e7d32', borderRadius: '30px' },
                '&:hover fieldset': { borderColor: '#4caf50' },
                '&.Mui-focused fieldset': { borderColor: '#4caf50' },
              },
            }}
          />
        </div>

        {/* Кнопки увеличения — появляются только после ввода имени */}
        {step === 'sample' && (
          <div className="magnification-buttons">
            {magnifications.map((mag) => (
              <button
                key={mag}
                data-mag={mag.toString()}
                className={`mag-btn ${magnification === mag ? 'active' : ''}`}
                onClick={() => setMagnification(mag)}
              >
                {mag}×
              </button>
            ))}
          </div>
        )}

        <Button
          className="send-btn"
          variant="contained"
          onClick={handleSend}
          disabled={isLoading}
        >
          Отправить
        </Button>
      </div>
    </div>
  );
};