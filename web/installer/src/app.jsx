import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import StepOne from "./stepOne";
import StepTwo from "./stepTwo";
import StepThree from "./stepThree";

const App = () => {
  const bootstrapMessage = {
    envName: "staging",
    infraRepo: "staging-infra",
    repoPerEnv: true,
    infraPublicKey: "key123",
    infraSecretFileName: "staging.yaml",
    infraGitopsRepoFileName: "staging.yaml",
    isNewInfraRepo: true,
    appsRepo: "staging-apps",
    appsPublicKey: "key123",
    appsSecretFileName: "staging.yaml",
    appsGitopsRepoFileName: "staging.yaml",

};

  return (
    <Router>

      <Switch>
        <Route exact path="/">
          <StepOne />
        </Route>

        <Route path="/step-2">
          <StepTwo
          appId={""}
          repoPerEnv={true}
          />
        </Route>

        <Route path="/step-3">
          <StepThree
          appId={""}
          infraRepo={"staging-infra"}
          appsRepo={"staging-apps"}
          bootstrapMessage={bootstrapMessage}
          />
        </Route>
      </Switch>
    </Router >
  );
};

export default App;
