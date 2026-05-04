import express from 'express';
import http from 'http';
import { WebSocketServer, WebSocket } from 'ws';
import cors from 'cors';
import axios from 'axios';

const PORT = 8001;
const HOST = 'localhost';
const AGENT_URL = 'http://localhost:8080';

const app = express();
const server = http.createServer(app);
const wss = new WebSocketServer({ server });

app.use(cors());
app.use(express.json());

const connections = new Map<string, WebSocket>();
let nextRequestId = 1;

app.post('/receive', (req, res) => {
  const message = req.body;
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
        ws.send(JSON.stringify({ type: 'accepted', message: 'Запрос принят...' }));
        
        const agentRequest = {
          request_id: nextRequestId++,
          sample_id: request.sample_id,
          magnification: request.magnification,
        };
        
        try {
          console.log(`[Agent] Отправка: sample=${request.sample_id}, mag=${request.magnification}`);
          const response = await axios.post(`${AGENT_URL}/send`, agentRequest);
          
          if (response.data.status === 'accepted' && response.data.images) {
            const wsResponse = {
              sender: request.sender,
              timestamp: new Date().toISOString(),
              error_flag: false,
              magnification: request.magnification,
              sample_id: request.sample_id,
              payload: { images: response.data.images },
            };
            ws.send(JSON.stringify({ type: 'response', data: wsResponse }));
            console.log(`[Agent] Отправлено ${response.data.images.length} изображений`);
          } else {
            ws.send(JSON.stringify({ type: 'error', message: response.data.error || 'Ошибка' }));
          }
        } catch (error: any) {
          console.error(`[Agent] Ошибка:`, error.message);
          ws.send(JSON.stringify({ type: 'error', message: 'Агент не отвечает' }));
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