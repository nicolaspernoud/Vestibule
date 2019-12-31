import * as Apps from "/components/apps/apps.js";
import * as Users from "/components/users/users.js";
import * as Login from "/components/login/login.js";
import * as Auth from "/services/auth/auth.js";
import { AnimateCSS } from "/services/common/common.js";

const mountPoint = document.getElementById("main");
const spinner = document.getElementById("spinner");

document.addEventListener("DOMContentLoaded", function() {
  // Hamburger menu
  const burger = document.getElementById("navbar-burger");
  const menu = document.getElementById("navbar-menu");
  burger.addEventListener("click", () => {
    burger.classList.toggle("is-active");
    menu.classList.toggle("is-active");
  });
  window.addEventListener("hashchange", navigate);
  location.hash = "#apps";
});

async function navigate() {
  await hideShowInterfaceElements();
  switch (location.hash) {
    case "#apps":
      load(mountPoint, async function() {
        await Apps.mount("main");
      });
      break;
    case "#users":
      load(mountPoint, async function() {
        await Users.mount("main");
      });
      break;
    case "#login":
      load(mountPoint, async function() {
        await Login.mount("main");
      });
      break;
    default:
      location.hash = "#apps";
      break;
  }
}

async function hideShowInterfaceElements() {
  const user = await Auth.GetUser();
  if (user === undefined) {
    document.getElementById("goto-users").classList.add("is-hidden");
    document.getElementById("goto-logout").classList.add("is-hidden");
    document.getElementById("goto-login").classList.remove("is-hidden");
  } else {
    document.getElementById("goto-logout").classList.remove("is-hidden");
    document.getElementById("goto-login").classList.add("is-hidden");
    if (user.isAdmin) {
      document.getElementById("goto-users").classList.remove("is-hidden");
    }
  }
}

async function load(element, domAlteration) {
  AnimateCSS(element, "zoomOut", async function() {
    element.classList.add("is-hidden");
    spinner.classList.remove("is-hidden");
    if (typeof domAlteration === "function") {
      await domAlteration();
      spinner.classList.add("is-hidden");
      element.classList.remove("is-hidden");
      AnimateCSS(element, "zoomIn");
    }
  });
}
