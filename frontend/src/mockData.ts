// src/mockData.ts
import { InputRequest, ReceiveMessage } from './types';

// Мок-изображение в base64 (маленький прозрачный квадрат, можно заменить на реальную картинку)
export const mockImageBase64 = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';

// Функция-заглушка для имитации ответа от транспортного уровня
// Через 2-3 секунды "возвращает" изображения
export const mockReceiveFromTransport = (
  request: InputRequest,
  requestId: number
): Promise<ReceiveMessage> => {
  return new Promise((resolve) => {
    // Имитируем задержку сети (1.5-3 секунды)
    const delay = 1500 + Math.random() * 1500;
    
    setTimeout(() => {
      // С вероятностью 10% имитируем ошибку (как в ТЗ с потерями)
      const shouldError = Math.random() < 0.1;
      
      // Генерируем 1-3 мок-изображения
      const imageCount = Math.floor(Math.random() * 3) + 1;
      const images = Array(imageCount).fill(mockImageBase64);
      
      const response: ReceiveMessage = {
        sender: request.sender,
        timestamp: new Date().toISOString(),
        error_flag: shouldError,
        magnification: request.magnification,
        sample_id: request.sample_id,
        payload: {
          images: shouldError ? [] : images,
        },
      };
      
      resolve(response);
    }, delay);
  });
};

// Функция для генерации текстового описания образца
export const getSampleDescription = (sampleId: number, magnification: number): string => {
  const descriptions: Record<number, string> = {
    1: 'Клетка растения n, ядро хорошо различимо',
    2: 'Митохондрии, поперечный срез',
    3: 'Клеточная стенка, утолщение',
    4: 'Хлоропласты, гранулы видны',
    5: 'Эпителиальная ткань, ядра окрашены',
  };
  
  return descriptions[sampleId] || `Образец №${sampleId}, увеличение ${magnification}×`;
};