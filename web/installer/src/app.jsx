import { BrowserRouter as Router, Route, Switch } from "react-router-dom";
import axios from 'axios';
import StepOne from "./stepOne";
import StepTwo from "./stepTwo";
import StepThree from "./stepThree";

const App = () => {
  const getContext = async () => {
    try {
      const resp = await axios.get('/context');
      return await resp.data;
    } catch (err) {
      // Handle Error Here
      console.error(`Error: ${err}`);
    }
  };

  return (
    <Router>

      <Switch>
        <Route exact path="/">
          <StepOne
            getContext={getContext}
          />
        </Route>

        <Route path="/step-2">
          <StepTwo
            getContext={getContext}
          />
        </Route>

        <Route path="/step-3">
          <StepThree
            getContext={getContext}
          />
        </Route>
      </Switch>
    </Router >
  );
};

export default App;
