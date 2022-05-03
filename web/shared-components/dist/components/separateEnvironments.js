"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;

var _react = _interopRequireDefault(require("react"));

var _react2 = require("@headlessui/react");

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

const SeparateEnvironments = _ref => {
  let {
    repoPerEnv,
    setRepoPerEnv,
    infraRepo,
    appsRepo,
    setInfraRepo,
    setAppsRepo
  } = _ref;
  return /*#__PURE__*/_react.default.createElement("div", {
    className: "text-gray-700"
  }, /*#__PURE__*/_react.default.createElement("div", {
    className: "flex mt-4"
  }, /*#__PURE__*/_react.default.createElement("div", {
    className: "font-medium self-center"
  }, "Separate environments by git repositories"), /*#__PURE__*/_react.default.createElement("div", {
    className: "max-w-lg flex rounded-md ml-4"
  }, /*#__PURE__*/_react.default.createElement(_react2.Switch, {
    checked: repoPerEnv,
    onChange: setRepoPerEnv,
    className: (repoPerEnv ? "bg-indigo-600" : "bg-gray-200") + " relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200"
  }, /*#__PURE__*/_react.default.createElement("span", {
    className: "sr-only"
  }, "Use setting"), /*#__PURE__*/_react.default.createElement("span", {
    "aria-hidden": "true",
    className: (repoPerEnv ? "translate-x-5" : "translate-x-0") + " pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200"
  })))), /*#__PURE__*/_react.default.createElement("div", {
    className: "text-sm text-gray-500 leading-loose"
  }, "Manifests will be placed in environment specific repositories"), /*#__PURE__*/_react.default.createElement("div", {
    className: "flex mt-4"
  }, /*#__PURE__*/_react.default.createElement("div", {
    className: "font-medium self-center"
  }, "Infrastructure git repository"), /*#__PURE__*/_react.default.createElement("div", {
    className: "max-w-lg flex rounded-md ml-4"
  }, /*#__PURE__*/_react.default.createElement("div", {
    className: "max-w-lg w-full lg:max-w-xs"
  }, /*#__PURE__*/_react.default.createElement("input", {
    id: "infra",
    name: "infra",
    className: "block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm",
    type: "text",
    value: infraRepo,
    onChange: e => setInfraRepo(e.target.value)
  })))), /*#__PURE__*/_react.default.createElement("div", {
    className: "text-sm text-gray-500 leading-loose"
  }, "Infrastructure manifests will be placed in the root of the specified repository"), /*#__PURE__*/_react.default.createElement("div", {
    className: "flex mt-4"
  }, /*#__PURE__*/_react.default.createElement("div", {
    className: "font-medium self-center"
  }, "Application git repository"), /*#__PURE__*/_react.default.createElement("div", {
    className: "max-w-lg flex rounded-md ml-4"
  }, /*#__PURE__*/_react.default.createElement("div", {
    className: "max-w-lg w-full lg:max-w-xs"
  }, /*#__PURE__*/_react.default.createElement("input", {
    id: "apps",
    name: "apps",
    className: "block w-full p-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm",
    type: "text",
    value: appsRepo,
    onChange: e => setAppsRepo(e.target.value)
  })))), /*#__PURE__*/_react.default.createElement("div", {
    className: "text-sm text-gray-500 leading-loose"
  }, "Application manifests will be placed in the root of the specified repository"));
};

var _default = SeparateEnvironments;
exports.default = _default;