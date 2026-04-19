# FileSystem.io Service

# Setup

## Amplify/React/Vite quick start
https://docs.amplify.aws/react/start/quickstart/


## Manual Process 

mkdir filesystem-service;cd filesystem-service
npm create amplify@latest
npx ampx sandbox
node errors -> 'nvm use 22'
npx ampx sandbox
npx ampx sandbox delete

//generate swift classes
npx ampx generate graphql-client-code --format modelgen --model-target swift


## Sandbox user management

npx ampx sandbox secret set username
npx ampx sandbox secret set password
npm install @aws-amplify/seed --save-dev
npx ampx sandbox seed 

## might have to delete sandbox: npx ampx sandbox delete


## Deploy React App

        - mkdir dist
        - cp amplify_outputs.json dist/amplify_outputs.json
        - cp amplify_outputs.json client/src/amplify_outputs.json
        - cd client
        - npm install
        - BUILD_PATH=../dist npm run build

## Amplify.yml
version: 1
backend:
  phases:
    build:
      commands:
        - npm ci --cache .npm --prefer-offline
        - npx ampx pipeline-deploy --branch $AWS_BRANCH --app-id $AWS_APP_ID
frontend:
  phases:
    build:
      commands:
        - mkdir dist
        - cp amplify_outputs.json dist/amplify_outputs.json
        - cd client
        - npm install
        - BUILD_PATH=../dist npm run build
  artifacts:
    baseDirectory: dist
    files:
      - '**/*'
  cache:
    paths:
      - .npm/**/*

## Start tracking via git

git init
git add README.md
git add package.json
git add .gitignore
git add amplify

