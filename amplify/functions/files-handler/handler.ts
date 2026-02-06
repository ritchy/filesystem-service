import type { APIGatewayProxyHandler } from 'aws-lambda';

export const handler: APIGatewayProxyHandler = async (event) => {
  console.log('event', event);

  const files = [
    {
      id: '/Code',
      date: new Date(2023, 11, 2, 17, 25),
      type: 'folder',
    },
    {
      id: '/Music',
      date: new Date(2023, 11, 1, 14, 45),
      type: 'folder',
    },
    {
      id: '/Music/Animal_sounds.mp3',
      size: 1457296,
      date: new Date(2023, 11, 1, 14, 45),
      type: 'file',
    }
  ];

  return {
    statusCode: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Headers': '*',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(files)
  };
};
