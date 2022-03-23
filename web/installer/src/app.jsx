import { BrowserRouter as Router, Redirect, Route, Switch } from "react-router-dom";
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
      <Route exact path="/">
        <Redirect to="/step-one" />
      </Route>

      <Switch>
        <Route path="/step-one">
          <StepOne />
        </Route>

        <Route path="/step-two">
          <StepTwo
          appId={""}
          repoPerEnv={true}
          />
        </Route>

        <Route path="/step-three">
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
