"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;

require("core-js/modules/es.string.includes.js");

require("core-js/modules/es.regexp.exec.js");

require("core-js/modules/es.string.split.js");

var _react = _interopRequireDefault(require("react"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

const BootstrapGuide = _ref => {
  let {
    envName,
    repoPath,
    repoPerEnv,
    publicKey,
    secretFileName,
    gitopsRepoFileName,
    isNewRepo
  } = _ref;
  const repoName = parseRepoName(repoPath);
  let type = "";

  if (repoPath.includes("apps")) {
    type = "apps";
  } else if (repoPath.includes("infra")) {
    type = "infra";
  }

  const renderBootstrapGuideText = isNewRepo => {
    return isNewRepo ? /*#__PURE__*/_react.default.createElement(_react.default.Fragment, null, /*#__PURE__*/_react.default.createElement("li", null, "\uD83D\uDC49 Clone the Gitops repository"), /*#__PURE__*/_react.default.createElement("li", {
      className: "text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded"
    }, "git clone git@github.com:", repoPath, ".git"), /*#__PURE__*/_react.default.createElement("li", null, "\uD83D\uDC49 Add the following deploy key to your Git provider"), /*#__PURE__*/_react.default.createElement("li", {
      className: "text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded"
    }, publicKey), /*#__PURE__*/_react.default.createElement("li", null, "\uD83D\uDC49 Apply the gitops manifests on the cluster to start the gitops loop:"), /*#__PURE__*/_react.default.createElement("ul", {
      className: "list-none text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded"
    }, /*#__PURE__*/_react.default.createElement("li", null, repoPerEnv ? "kubectl apply -f ".concat(repoName, "/flux/flux.yaml") : "kubectl apply -f ".concat(repoName, "/").concat(envName, "/flux/flux.yaml")), /*#__PURE__*/_react.default.createElement("li", null, repoPerEnv ? "kubectl apply -f ".concat(repoName, "/flux/").concat(secretFileName) : "kubectl apply -f ".concat(repoName, "/").concat(envName, "/flux/").concat(secretFileName)), /*#__PURE__*/_react.default.createElement("li", null, "kubectl wait --for condition=established --timeout=60s crd/gitrepositories.source.toolkit.fluxcd.io"), /*#__PURE__*/_react.default.createElement("li", null, "kubectl wait --for condition=established --timeout=60s crd/kustomizations.kustomize.toolkit.fluxcd.io"), /*#__PURE__*/_react.default.createElement("li", null, repoPerEnv ? "kubectl apply -f ".concat(repoName, "/flux/").concat(gitopsRepoFileName) : "kubectl apply -f ".concat(repoName, "/").concat(envName, "/flux/").concat(gitopsRepoFileName))), /*#__PURE__*/_react.default.createElement("li", null, "Happy Gitopsing\uD83C\uDF8A")) : /*#__PURE__*/_react.default.createElement(_react.default.Fragment, null, /*#__PURE__*/_react.default.createElement("li", null, "\uD83D\uDC49 Clone the Gitops repository"), /*#__PURE__*/_react.default.createElement("li", {
      className: "text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded"
    }, "git clone git@github.com:", repoPath, ".git"), /*#__PURE__*/_react.default.createElement("li", null, "\uD83D\uDC49 Apply the gitops manifests on the cluster to start the gitops loop:"), /*#__PURE__*/_react.default.createElement("ul", {
      className: "list-none text-xs font-mono bg-blue-100 font-medium text-blue-500 px-1 py-1 rounded"
    }, /*#__PURE__*/_react.default.createElement("li", null, repoPerEnv ? "kubectl apply -f ".concat(repoName, "/flux/").concat(gitopsRepoFileName) : "kubectl apply -f ".concat(repoName, "/").concat(envName, "/flux/").concat(gitopsRepoFileName))), /*#__PURE__*/_react.default.createElement("li", null, "Happy Gitopsing\uD83C\uDF8A"));
  };

  return /*#__PURE__*/_react.default.createElement("div", {
    className: "rounded-md bg-blue-50 p-4 mb-4 overflow-hidden"
  }, /*#__PURE__*/_react.default.createElement("ul", {
    className: "break-all text-sm text-blue-700 space-y-2"
  }, /*#__PURE__*/_react.default.createElement("span", {
    className: "text-lg font-bold text-blue-800"
  }, "Gitops ", type), renderBootstrapGuideText(isNewRepo)));
};

const parseRepoName = repo => {
  return repo.split("/")[1];
};

var _default = BootstrapGuide;
exports.default = _default;