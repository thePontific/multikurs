// src/types/index.ts

// Тип для запроса от прикладного уровня к транспортному (POST /input)
export interface InputRequest {
  sender: string;        // имя отправителя
  timestamp: string;     // ISO формат: "2024-03-27T10:30:00Z"
  magnification: number; // 10, 40, 100, 400, 1000
  sample_id: number;     // номер образца
}

// Тип для ответа от транспортного уровня (POST /input response)
export interface InputResponse {
  status: 'accepted' | 'error';
  request_id: number;    // уникальный ID запроса
}

// Тип для сообщения от транспортного уровня на прикладной (POST /receive)
export interface ReceiveMessage {
  sender: string;
  timestamp: string;
  error_flag: boolean;
  magnification: number;
  sample_id: number;
  payload: {
    images: string[];    // массив изображений в base64
  };
}

// Тип для сообщения в чате (для отображения)
export interface ChatMessage {
  id: string;
  sender: string;
  timestamp: string;
  isOwn: boolean;        // моё сообщение или ответ системы
  error_flag: boolean;
  magnification: number;
  sample_id: number;
  images?: string[];     // base64 изображения
  text?: string;         // текстовое описание
}

// Тип для WebSocket соединения
export interface WebSocketMessage {
  type: 'request' | 'response' | 'error';
  data: InputRequest | ReceiveMessage;
}