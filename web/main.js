import * as Apps from "/components/apps/apps.js";
import * as Users from "/components/users/users.js";
import * as Login from "/components/login/login.js";
import * as Auth from "/services/auth/auth.js";

document.addEventListener("DOMContentLoaded", function() {
  // Hamburger menu
  const burger = document.getElementById("navbar-burger");
  const menu = document.getElementById("navbar-menu");
  burger.addEventListener("click", () => {
    burger.classList.toggle("is-active");
    menu.classList.toggle("is-active");
  });

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
  if (user != undefined) {
    document.getElementById("goto-logout").classList.toggle("is-hidden");
    document.getElementById("goto-login").classList.toggle("is-hidden");
  }
}
