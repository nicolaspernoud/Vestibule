import * as Apps from "/apps/apps.js";
import * as Users from "/users/users.js";
import * as Login from "/login/login.js";

init();

function init() {
  document.getElementById("goto-apps").addEventListener("click", function() {
    Apps.mount("main");
  });
  document.getElementById("goto-users").addEventListener("click", function() {
    Users.mount("main");
  });
  document.getElementById("goto-login").addEventListener("click", function() {
    Login.mount("main");
  });
  Apps.mount("main");
}
