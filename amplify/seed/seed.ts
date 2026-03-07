import { readFile } from "node:fs/promises";
import {
  addToUserGroup,
  createAndSignUpUser,
  getSecret,
} from "@aws-amplify/seed";
import { Amplify } from "aws-amplify";

// this is used to get the amplify_outputs.json file as the file will not exist until sandbox is created
const url = new URL("../../amplify_outputs.json", import.meta.url);
const outputs = JSON.parse(await readFile(url, { encoding: "utf8" }));
Amplify.configure(outputs);

const username = await getSecret("username");
const password = await getSecret("password");

const user = await createAndSignUpUser({
  username: username,
  password: password,
  signInAfterCreation: false,
  signInFlow: "Password",
  userAttributes: {
    locale: "en",
  },
});

//await addToUserGroup(user, "admin");
