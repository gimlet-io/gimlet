"use strict";

require("core-js/modules/web.dom-collections.iterator.js");

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;

require("core-js/modules/es.symbol.description.js");

var _react = _interopRequireWildcard(require("react"));

require("./style.css");

var _category = require("./category");

function _getRequireWildcardCache(nodeInterop) { if (typeof WeakMap !== "function") return null; var cacheBabelInterop = new WeakMap(); var cacheNodeInterop = new WeakMap(); return (_getRequireWildcardCache = function _getRequireWildcardCache(nodeInterop) { return nodeInterop ? cacheNodeInterop : cacheBabelInterop; })(nodeInterop); }

function _interopRequireWildcard(obj, nodeInterop) { if (!nodeInterop && obj && obj.__esModule) { return obj; } if (obj === null || typeof obj !== "object" && typeof obj !== "function") { return { default: obj }; } var cache = _getRequireWildcardCache(nodeInterop); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (key !== "default" && Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj.default = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

class StackUI extends _react.Component {
  constructor(props) {
    super(props);
    this.state = {
      showErrors: false
    };
  }

  render() {
    let {
      stack,
      stackDefinition,
      setValues,
      validationCallback,
      categoriesToRender,
      componentsToRender,
      hideTitle
    } = this.props;

    if (stackDefinition === undefined || stack === undefined) {
      return null;
    }

    const categories = stackDefinition.categories.map(category => {
      if (categoriesToRender) {
        const toRender = categoriesToRender.find(c => category.id === c);

        if (!toRender) {
          return null;
        }
      }

      return /*#__PURE__*/_react.default.createElement(_category.Category, {
        category: category,
        stackDefinition: stackDefinition,
        stack: stack,
        genericComponentSaver: setValues,
        genericValidationCallback: validationCallback,
        componentsToRender: componentsToRender
      });
    });
    return /*#__PURE__*/_react.default.createElement("div", null, /*#__PURE__*/_react.default.createElement("div", null, /*#__PURE__*/_react.default.createElement("h1", {
      className: hideTitle ? "hidden" : "text-2xl font-bold mb-4"
    }, stackDefinition.name, /*#__PURE__*/_react.default.createElement("span", {
      className: "font-normal text-lg block"
    }, stackDefinition.description)), categories));
  }

}

;
var _default = StackUI;
exports.default = _default;