import type { APIGatewayProxyHandler } from 'aws-lambda';

// Function to generate a text file SVG icon
function generateTextIcon(size: 'big' | 'small'): string {
  const dimensions = size === 'big' ? { width: 64, height: 64, fontSize: 24 } : { width: 24, height: 24, fontSize: 10 };
  
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${dimensions.width}" height="${dimensions.height}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
  <polyline points="14 2 14 8 20 8"/>
  <line x1="16" y1="13" x2="8" y2="13"/>
  <line x1="16" y1="17" x2="8" y2="17"/>
  <polyline points="10 9 9 9 8 9"/>
</svg>`;
}

export const handler: APIGatewayProxyHandler = async (event) => {
  console.log('Icons handler event:', event);

  try {
    // Get size and name from path parameters
    const size = event.pathParameters?.size as 'big' | 'small' | undefined;
    const name = event.pathParameters?.name;

    if (!size || !name) {
      return {
        statusCode: 400,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ error: 'Missing size or name parameter in path' }),
      };
    }

    // Validate size parameter
    if (size !== 'big' && size !== 'small') {
      return {
        statusCode: 400,
        headers: {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Headers': '*',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ error: 'Size must be either "big" or "small"' }),
      };
    }

    // Generate the SVG icon (for now, we generate a text icon for all requests)
    const svgContent = generateTextIcon(size);

    return {
      statusCode: 200,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': '*',
        'Content-Type': 'image/svg+xml',
        //'Cache-Control': 'public, max-age=3600',
      },
      body: svgContent,
    };
  } catch (error) {
    console.error('Error generating icon:', error);
    return {
      statusCode: 500,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Headers': '*',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        error: 'Internal server error',
        message: error instanceof Error ? error.message : 'Unknown error',
      }),
    };
  }
};
