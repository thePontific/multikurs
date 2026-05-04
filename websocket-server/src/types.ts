// types.ts

// Запрос от фронта через WebSocket
export interface WebSocketRequest {
  type: 'send_request';
  sender: string;
  timestamp: string;
  magnification: number;
  sample_id: number;
}

// Запрос к транспортному уровню (POST /input)
export interface TransportInputRequest {
  sender: string;
  timestamp: string;
  magnification: number;
  sample_id: number;
}

// Ответ от транспортного уровня (POST /receive)
export interface TransportReceiveMessage {
  sender: string;
  timestamp: string;
  error_flag: boolean;
  magnification: number;
  sample_id: number;
  payload: {
    images: string[];
  };
}

// Сообщение для отправки фронту через WebSocket
export interface WSMessageToClient {
  type: 'response' | 'error';
  data: TransportReceiveMessage;
}