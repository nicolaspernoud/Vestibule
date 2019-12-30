import * as Apps from "/components/apps/apps.js";
import * as Users from "/components/users/users.js";
import * as Login from "/components/login/login.js";
import * as Auth from "/services/auth/auth.js";

document.addEventListener("DOMContentLoaded", function() {
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
  showToAdminsOnly();
});

async function showToAdminsOnly() {
  const user = await Auth.GetUser();
  if (user.isAdmin) {
    document.getElementById("goto-users").classList.toggle("is-hidden");
  }
}
