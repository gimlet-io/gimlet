"use strict";

require("core-js/modules/web.dom-collections.iterator.js");

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.Tile = void 0;

require("./style.css");

var _react = _interopRequireWildcard(require("react"));

function _getRequireWildcardCache(nodeInterop) { if (typeof WeakMap !== "function") return null; var cacheBabelInterop = new WeakMap(); var cacheNodeInterop = new WeakMap(); return (_getRequireWildcardCache = function _getRequireWildcardCache(nodeInterop) { return nodeInterop ? cacheNodeInterop : cacheBabelInterop; })(nodeInterop); }

function _interopRequireWildcard(obj, nodeInterop) { if (!nodeInterop && obj && obj.__esModule) { return obj; } if (obj === null || typeof obj !== "object" && typeof obj !== "function") { return { default: obj }; } var cache = _getRequireWildcardCache(nodeInterop); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (key !== "default" && Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj.default = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

class Tile extends _react.Component {
  constructor(props) {
    super(props);
  }

  render() {
    const {
      category,
      component,
      componentConfig,
      selectedComponentName,
      toggleComponentHandler
    } = this.props;
    const selected = component.variable === selectedComponentName;
    const enabled = componentConfig !== undefined ? componentConfig.enabled : false;
    return /*#__PURE__*/_react.default.createElement("div", {
      onClick: () => toggleComponentHandler(category.id, component.variable)
    }, /*#__PURE__*/_react.default.createElement("div", {
      className: "w-32 h-32 px-2 overflow-hidden cursor-pointer"
    }, /*#__PURE__*/_react.default.createElement("div", {
      className: enabled ? !selected ? 'bg-green-100 hover:bg-green-300' : 'bg-green-300' : !selected ? 'bg-gray-200 hover:bg-gray-300 filter grayscale hover:grayscale-0' : 'bg-gray-300'
    }, /*#__PURE__*/_react.default.createElement("img", {
      className: "h-20 mx-auto pt-4",
      src: component.logo,
      alt: component.name
    }), /*#__PURE__*/_react.default.createElement("div", {
      className: "font-bold text-sm py-2 text-center"
    }, component.name))));
  }

}

exports.Tile = Tile;