"use strict";

require("core-js/modules/web.dom-collections.iterator.js");

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;

require("core-js/modules/es.promise.js");

require("core-js/modules/es.json.stringify.js");

require("core-js/modules/es.symbol.description.js");

var _react = _interopRequireWildcard(require("react"));

var _reactHotLoader = require("react-hot-loader");

var stackDefinitionFixture = _interopRequireWildcard(require("../fixtures/stack-definition.json"));

require("./style.css");

var _streamingBackend = _interopRequireDefault(require("./streamingBackend"));

var _client = _interopRequireDefault(require("./client"));

var _category = require("./components/category");

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _getRequireWildcardCache(nodeInterop) { if (typeof WeakMap !== "function") return null; var cacheBabelInterop = new WeakMap(); var cacheNodeInterop = new WeakMap(); return (_getRequireWildcardCache = function _getRequireWildcardCache(nodeInterop) { return nodeInterop ? cacheNodeInterop : cacheBabelInterop; })(nodeInterop); }

function _interopRequireWildcard(obj, nodeInterop) { if (!nodeInterop && obj && obj.__esModule) { return obj; } if (obj === null || typeof obj !== "object" && typeof obj !== "function") { return { default: obj }; } var cache = _getRequireWildcardCache(nodeInterop); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (key !== "default" && Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj.default = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function ownKeys(object, enumerableOnly) { var keys = Object.keys(object); if (Object.getOwnPropertySymbols) { var symbols = Object.getOwnPropertySymbols(object); enumerableOnly && (symbols = symbols.filter(function (sym) { return Object.getOwnPropertyDescriptor(object, sym).enumerable; })), keys.push.apply(keys, symbols); } return keys; }

function _objectSpread(target) { for (var i = 1; i < arguments.length; i++) { var source = null != arguments[i] ? arguments[i] : {}; i % 2 ? ownKeys(Object(source), !0).forEach(function (key) { _defineProperty(target, key, source[key]); }) : Object.getOwnPropertyDescriptors ? Object.defineProperties(target, Object.getOwnPropertyDescriptors(source)) : ownKeys(Object(source)).forEach(function (key) { Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key)); }); } return target; }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

class App extends _react.Component {
  constructor(props) {
    super(props);
    const client = new _client.default();

    client.onError = response => {
      console.log(response);
      console.log("".concat(response.status, ": ").concat(response.statusText, " on ").concat(response.path));
    };

    this.state = {
      client: client,
      stack: {},
      stackNonDefaultValues: {},
      errors: {},
      showErrors: false
    };
    this.setValues = this.setValues.bind(this);
    this.validationCallback = this.validationCallback.bind(this);
  }

  componentDidMount() {
    fetch('/stack-definition.json').then(response => {
      if (!response.ok && window !== undefined) {
        console.log("Using fixture");
        return stackDefinitionFixture.default;
      }

      return response.json();
    }).then(data => this.setState({
      stackDefinition: data
    }));
    fetch('/stack.json').then(response => {
      if (!response.ok && window !== undefined) {
        console.log("Using fixture");
        return {};
      }

      return response.json();
    }).then(data => this.setState({
      stack: data
    }));
  }

  setValues(variable, values, nonDefaultValues) {
    const updatedNonDefaultValues = _objectSpread(_objectSpread({}, this.state.stackNonDefaultValues), {}, {
      [variable]: nonDefaultValues
    });

    this.setState(prevState => ({
      stack: _objectSpread(_objectSpread({}, prevState.stack), {}, {
        [variable]: values
      }),
      stackNonDefaultValues: _objectSpread(_objectSpread({}, prevState.stackNonDefaultValues), {}, {
        [variable]: nonDefaultValues
      })
    }));
  }

  validationCallback(variable, errors) {
    if (errors === null) {
      this.setState(prevState => {
        delete prevState.errors[variable];

        if (JSON.stringify(prevState.errors) === "{}") {
          return {
            errors: {},
            showErrors: false
          };
        }

        return {
          errors: prevState.errors
        };
      });
      return;
    }

    errors = errors.filter(error => error.keyword !== 'oneOf');
    errors = errors.filter(error => error.dataPath !== '.enabled');
    this.setState(prevState => ({
      errors: _objectSpread(_objectSpread({}, prevState.errors), {}, {
        [variable]: errors
      })
    }));
  }

  render() {
    let {
      stackDefinition,
      stack
    } = this.state;

    if (stackDefinition === undefined || stack === undefined) {
      return null;
    }

    const categories = stackDefinition.categories.map(category => {
      return /*#__PURE__*/_react.default.createElement(_category.Category, {
        category: category,
        stackDefinition: stackDefinition,
        stack: stack,
        genericComponentSaver: this.setValues,
        genericValidationCallback: this.validationCallback
      });
    });
    return /*#__PURE__*/_react.default.createElement("div", null, /*#__PURE__*/_react.default.createElement(_streamingBackend.default, {
      client: this.state.client
    }), /*#__PURE__*/_react.default.createElement("div", {
      className: this.state.showErrors ? 'block fixed bottom-0 right-0 mb-48 mr-8 bg-red-300 rounded-md shadow py-4 px-8' : 'hidden'
    }, /*#__PURE__*/_react.default.createElement("ul", {
      class: "list-disc list-inside"
    }, Object.keys(this.state.errors).map(variable => {
      return /*#__PURE__*/_react.default.createElement("div", null, /*#__PURE__*/_react.default.createElement("p", {
        class: "capitalize font-bold"
      }, variable), this.state.errors[variable].map(e => {
        return /*#__PURE__*/_react.default.createElement("li", null, e.message);
      }));
    }))), /*#__PURE__*/_react.default.createElement("div", {
      className: "fixed bottom-0 right-0"
    }, /*#__PURE__*/_react.default.createElement("span", {
      className: "inline-flex rounded-md shadow-sm m-8"
    }, /*#__PURE__*/_react.default.createElement("button", {
      type: "button",
      className: "inline-flex items-center px-12 py-6 border border-transparent text-base font-medium rounded-md shadow-sm text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500",
      onClick: () => {
        console.log(this.state.stack);
        console.log(this.state.stackNonDefaultValues);

        if (JSON.stringify(this.state.errors) !== "{}") {
          this.setState(() => ({
            showErrors: true
          }));
          return false;
        }

        this.state.client.saveValues(this.state.stackNonDefaultValues).then(() => {
          close();
        });
      }
    }, "Close tab & ", /*#__PURE__*/_react.default.createElement("br", null), "Write config"))), /*#__PURE__*/_react.default.createElement("div", {
      className: "container mx-auto m-8 max-w-4xl"
    }, /*#__PURE__*/_react.default.createElement("h1", {
      class: "text-2xl font-bold my-16"
    }, stackDefinition.name, /*#__PURE__*/_react.default.createElement("span", {
      class: "font-normal text-lg block"
    }, stackDefinition.description)), categories));
  }

}

;

var _default = (0, _reactHotLoader.hot)(module)(App);

exports.default = _default;