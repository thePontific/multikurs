import axios from 'axios';
import { TransportInputRequest, TransportReceiveMessage } from '../types';

// TODO: позже замените на реальный IP транспортного уровня
const TRANSPORT_URL = 'http://localhost:3030'; // транспортный уровень (порт из Swagger)

// Отправка запроса на транспортный уровень (POST /input)
export const sendToTransport = async (request: TransportInputRequest): Promise<{ request_id: number }> => {
  try {
    const response = await axios.post(`${TRANSPORT_URL}/input`, request);
    return { request_id: response.data.request_id };
  } catch (error) {
    console.error('Ошибка отправки на транспортный уровень:', error);
    throw new Error('Транспортный уровень недоступен');
  }
};