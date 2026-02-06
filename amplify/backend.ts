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

// Add data environment variables to the function
backend.filesHandler.addEnvironment('AMPLIFY_DATA_GRAPHQL_ENDPOINT', backend.data.resources.cfnResources.cfnGraphqlApi.attrGraphQlUrl);

// Grant Lambda access to AppSync
import { PolicyStatement, Effect } from 'aws-cdk-lib/aws-iam';

backend.filesHandler.resources.lambda.addToRolePolicy(
  new PolicyStatement({
    effect: Effect.ALLOW,
    actions: ['appsync:GraphQL'],
    resources: [`${backend.data.resources.cfnResources.cfnGraphqlApi.attrArn}/*`],
  })
);

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

// Add /info endpoint
const infoResource = apiGateway.root.addResource('info');
infoResource.addMethod('GET', lambdaIntegration);

// Add /direct endpoint
const directResource = apiGateway.root.addResource('direct');
directResource.addMethod('GET', lambdaIntegration);

// Add outputs
backend.addOutput({
  custom: {
    FilesApiUrl: apiGateway.url,
    FilesApiId: apiGateway.restApiId,
  },
});
