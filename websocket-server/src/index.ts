import express from 'express';
import http from 'http';
import { WebSocketServer, WebSocket } from 'ws';
import cors from 'cors';
import axios from 'axios';

const PORT = 8001;
const HOST = '0.0.0.0';
//const TRANSPORT_URL = 'http://localhost:8000';  // Транспортный уровень 
const TRANSPORT_URL = 'http://transport:8000';
const app = express();
const server = http.createServer(app);
const wss = new WebSocketServer({ server });

app.use(cors());
app.use(express.json({ limit: '50mb' }));

const connections = new Map<string, WebSocket>();

// Метод /receive - сюда транспортный уровень отправляет собранные сообщения
app.post('/receive', (req, res) => {
  const message = req.body;
  console.log(`[HTTP /receive] Получено сообщение для ${message.sender}`);
  const clientWs = connections.get(message.sender);
  if (clientWs && clientWs.readyState === WebSocket.OPEN) {
    clientWs.send(JSON.stringify({ type: 'response', data: message }));
    res.status(200).json({ status: 'ok' });
  } else {
    res.status(404).json({ status: 'error', message: 'User not connected' });
  }
});

wss.on('connection', (ws: WebSocket, req: http.IncomingMessage) => {
  const url = new URL(req.url || '', `http://${req.headers.host}`);
  const username = url.searchParams.get('username');
  
  if (!username) {
    ws.close(1008, 'Username required');
    return;
  }
  
  connections.set(username, ws);
  console.log(`[WebSocket] Пользователь ${username} подключен`);
  ws.send(JSON.stringify({ type: 'connected', message: 'Добро пожаловать' }));
  
  ws.on('message', async (data: string) => {
    try {
      const request = JSON.parse(data);
      console.log(`[WebSocket] Запрос от ${username}:`, request);
      
      if (request.type === 'send_request') {
        ws.send(JSON.stringify({ type: 'accepted', message: 'Запрос принят, ожидайте ответа...' }));
        
        // Отправляем запрос в ТРАНСПОРТНЫЙ уровень (НЕ напрямую агенту)
        const transportRequest = {
          sender: request.sender,
          timestamp: request.timestamp,
          magnification: request.magnification,
          sample_id: String(request.sample_id),  // ← ПРЕОБРАЗУЕМ В СТРОКУ
        };
        
// Найти эту секцию и заменить
      try {
        console.log(`[Transport] Отправка в транспортный уровень:`, JSON.stringify(transportRequest, null, 2));
        const transportResponse = await axios.post(`${TRANSPORT_URL}/input`, transportRequest, {
          headers: { 'Content-Type': 'application/json' }
        });
        console.log(`[Transport] Ответ:`, transportResponse.data);
      } catch (error: any) {
        console.error(`[Transport] Ошибка:`, error.response?.data || error.message);
        console.error(`[Transport] Статус:`, error.response?.status);
        ws.send(JSON.stringify({ type: 'error', message: 'Транспортный уровень недоступен' }));
      }
      }
    } catch (error) {
      console.error('[WebSocket] Ошибка:', error);
      ws.send(JSON.stringify({ type: 'error', message: 'Неверный формат' }));
    }
  });
  
  ws.on('close', () => {
    connections.delete(username);
    console.log(`[WebSocket] ${username} отключен`);
  });
});

server.listen(PORT, HOST, () => {
  console.log(`WebSocket-сервер запущен на ws://${HOST}:${PORT}`);
});