import { defineBackend } from '@aws-amplify/backend';
import { auth } from './auth/resource';
import { data } from './data/resource';
import { filesHandler } from './functions/files-handler/resource';

/**
 * @see https://docs.amplify.aws/react/build-a-backend/ to add storage, functions, and more
 */
const backend = defineBackend({
  auth,
  data,
  filesHandler,
});

// Create API Gateway REST API
const filesApi = backend.createStack('files-api');

// Import API Gateway v1 (REST API) construct
import { RestApi, LambdaIntegration, Cors } from 'aws-cdk-lib/aws-apigateway';

const apiGateway = new RestApi(filesApi, 'FilesRestApi', {
  restApiName: 'Files API',
  description: 'API for filesystem operations',
  deployOptions: {
    stageName: 'dev',
  },
  defaultCorsPreflightOptions: {
    allowOrigins: Cors.ALL_ORIGINS,
    allowMethods: Cors.ALL_METHODS,
    allowHeaders: ['*'],
  },
});

// Create Lambda integration
const lambdaIntegration = new LambdaIntegration(
  backend.filesHandler.resources.lambda
);

// Add /files endpoint
const filesResource = apiGateway.root.addResource('files');
filesResource.addMethod('GET', lambdaIntegration);

// Add outputs
backend.addOutput({
  custom: {
    FilesApiUrl: apiGateway.url,
    FilesApiId: apiGateway.restApiId,
  },
});
