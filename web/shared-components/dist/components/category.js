"use strict";

require("core-js/modules/web.dom-collections.iterator.js");

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.Category = void 0;

require("./style.css");

var _react = _interopRequireWildcard(require("react"));

var _tile = require("./tile");

var _helmReactUi = _interopRequireDefault(require("helm-react-ui"));

var _outline = require("@heroicons/react/outline");

var _remarkable = require("remarkable");

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _getRequireWildcardCache(nodeInterop) { if (typeof WeakMap !== "function") return null; var cacheBabelInterop = new WeakMap(); var cacheNodeInterop = new WeakMap(); return (_getRequireWildcardCache = function _getRequireWildcardCache(nodeInterop) { return nodeInterop ? cacheNodeInterop : cacheBabelInterop; })(nodeInterop); }

function _interopRequireWildcard(obj, nodeInterop) { if (!nodeInterop && obj && obj.__esModule) { return obj; } if (obj === null || typeof obj !== "object" && typeof obj !== "function") { return { default: obj }; } var cache = _getRequireWildcardCache(nodeInterop); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (key !== "default" && Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj.default = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function ownKeys(object, enumerableOnly) { var keys = Object.keys(object); if (Object.getOwnPropertySymbols) { var symbols = Object.getOwnPropertySymbols(object); enumerableOnly && (symbols = symbols.filter(function (sym) { return Object.getOwnPropertyDescriptor(object, sym).enumerable; })), keys.push.apply(keys, symbols); } return keys; }

function _objectSpread(target) { for (var i = 1; i < arguments.length; i++) { var source = null != arguments[i] ? arguments[i] : {}; i % 2 ? ownKeys(Object(source), !0).forEach(function (key) { _defineProperty(target, key, source[key]); }) : Object.getOwnPropertyDescriptors ? Object.defineProperties(target, Object.getOwnPropertyDescriptors(source)) : ownKeys(Object(source)).forEach(function (key) { Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key)); }); } return target; }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

class Category extends _react.Component {
  constructor(props) {
    super(props);
    this.state = {
      toggleState: {},
      tabState: {}
    };
    this.toggleComponent = this.toggleComponent.bind(this);
  }

  toggleComponent(category, component) {
    this.setState(prevState => ({
      toggleState: _objectSpread(_objectSpread({}, prevState.toggleState), {}, {
        [category]: prevState.toggleState[category] == component ? undefined : component
      }),
      tabState: _objectSpread(_objectSpread({}, prevState.tabState), {}, {
        [component]: prevState.tabState[component] === undefined ? 'getting-started' : prevState.tabState[component]
      })
    }));
  }

  switchTab(component, tab) {
    this.setState(prevState => ({
      tabState: _objectSpread(_objectSpread({}, prevState.tabState), {}, {
        [component]: tab
      })
    }));
  }

  render() {
    let {
      toggleState
    } = this.state;
    const {
      category,
      stackDefinition,
      stack,
      genericComponentSaver,
      genericValidationCallback
    } = this.props;
    let selectedComponent = undefined;
    let selectedComponentConfig = undefined;
    let componentSaver = undefined;
    let validationCallback = undefined;
    const selectedComponentName = toggleState[category.id];

    if (selectedComponentName !== undefined) {
      const selectedComponentArray = stackDefinition.components.filter(component => component.variable === toggleState[category.id]);
      selectedComponent = selectedComponentArray[0];
      selectedComponentConfig = stack[selectedComponent.variable];

      if (selectedComponentConfig === undefined) {
        selectedComponentConfig = {};
      }

      componentSaver = function componentSaver(values, nonDefaultValues) {
        genericComponentSaver(selectedComponent.variable, values, nonDefaultValues);
      };

      validationCallback = function validationCallback(errors) {
        genericValidationCallback(selectedComponent.variable, errors);
      };
    }

    const componentsForCategory = stackDefinition.components.filter(component => component.category === category.id);
    const componentTitles = componentsForCategory.map(component => {
      return /*#__PURE__*/_react.default.createElement(_tile.Tile, {
        category: category,
        component: component,
        componentConfig: stack[component.variable],
        selectedComponentName: selectedComponentName,
        toggleComponentHandler: this.toggleComponent
      });
    });

    if (selectedComponentName !== undefined) {
      if (typeof selectedComponent.schema !== 'object') {
        selectedComponent.schema = JSON.parse(selectedComponent.schema);
      }

      if (typeof selectedComponent.uiSchema !== 'object') {
        selectedComponent.uiSchema = JSON.parse(selectedComponent.uiSchema);
      }
    }

    const componentConfigPanel = selectedComponentName === undefined ? null : /*#__PURE__*/_react.default.createElement("div", {
      className: "py-6 px-4 space-y-6 sm:p-6"
    }, /*#__PURE__*/_react.default.createElement(_helmReactUi.default, {
      schema: selectedComponent.schema,
      config: selectedComponent.uiSchema,
      values: selectedComponentConfig,
      setValues: componentSaver,
      validate: true,
      validationCallback: validationCallback
    }));
    const md = new _remarkable.Remarkable();
    const gettingStartedPanel = selectedComponentName === undefined ? null : /*#__PURE__*/_react.default.createElement("div", {
      className: "py-6 px-4 space-y-6 sm:p-6"
    }, /*#__PURE__*/_react.default.createElement("div", {
      class: "prose",
      dangerouslySetInnerHTML: {
        __html: md.render(selectedComponent.onePager)
      }
    }));
    const notSelectedTabStyle = "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm";
    const selectedTabStyle = "border-indigo-500 text-indigo-600 whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm";
    const tabState = this.state.tabState;
    return /*#__PURE__*/_react.default.createElement("div", {
      class: "my-8"
    }, /*#__PURE__*/_react.default.createElement("h2", {
      class: "text-lg"
    }, category.name), /*#__PURE__*/_react.default.createElement("div", {
      className: "flex space-x-2 my-2"
    }, componentTitles), /*#__PURE__*/_react.default.createElement("div", {
      className: "my-2"
    }, selectedComponentName !== undefined && /*#__PURE__*/_react.default.createElement("div", {
      className: "px-8 py-4 shadow sm:rounded-md sm:overflow-hidden bg-white relative"
    }, /*#__PURE__*/_react.default.createElement("div", {
      className: "hidden sm:block absolute top-0 right-0 pt-4 pr-4"
    }, /*#__PURE__*/_react.default.createElement("button", {
      type: "button",
      className: "bg-white rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500",
      onClick: () => this.toggleComponent(category.id, selectedComponent.variable)
    }, /*#__PURE__*/_react.default.createElement("span", {
      className: "sr-only"
    }, "Close"), /*#__PURE__*/_react.default.createElement(_outline.XIcon, {
      className: "h-6 w-6",
      "aria-hidden": "true"
    }))), /*#__PURE__*/_react.default.createElement("div", null, /*#__PURE__*/_react.default.createElement("div", {
      className: "sm:hidden"
    }, /*#__PURE__*/_react.default.createElement("label", {
      htmlFor: "tabs",
      className: "sr-only"
    }, "Select a tab"), /*#__PURE__*/_react.default.createElement("select", {
      id: "tabs",
      name: "tabs",
      className: "block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
    }, /*#__PURE__*/_react.default.createElement("option", null, "Getting Started"), /*#__PURE__*/_react.default.createElement("option", {
      selected: true
    }, "Config"))), /*#__PURE__*/_react.default.createElement("div", {
      className: "hidden sm:block"
    }, /*#__PURE__*/_react.default.createElement("div", {
      className: "border-b border-gray-200"
    }, /*#__PURE__*/_react.default.createElement("nav", {
      className: "-mb-px flex space-x-8",
      "aria-label": "Tabs"
    }, /*#__PURE__*/_react.default.createElement("button", {
      className: tabState[selectedComponentName] == 'getting-started' ? selectedTabStyle : notSelectedTabStyle,
      "aria-current": tabState[selectedComponentName] == 'getting-started' ? 'page' : undefined,
      onClick: () => this.switchTab(selectedComponentName, 'getting-started')
    }, "Getting Started"), /*#__PURE__*/_react.default.createElement("button", {
      className: tabState[selectedComponentName] == 'config' ? selectedTabStyle : notSelectedTabStyle,
      "aria-current": tabState[selectedComponentName] == 'config' ? 'page' : undefined,
      onClick: () => this.switchTab(selectedComponentName, 'config')
    }, "Config"))))), tabState[selectedComponentName] == 'getting-started' && gettingStartedPanel, tabState[selectedComponentName] == 'config' && componentConfigPanel)));
  }

}

exports.Category = Category;