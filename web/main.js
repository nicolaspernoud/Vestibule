import * as Apps from "/components/apps/apps.js";
import * as Davs from "/components/davs/davs.js";
import * as Users from "/components/users/users.js";
import { Login } from "/components/login/login.js";
import * as Sysinfo from "/components/sysinfo/sysinfo.js";
import { Navbar } from "/components/navbar/navbar.js";
import { AnimateCSS } from "/services/common/common.js";
import * as Auth from "/services/auth/auth.js";

const mountPoint = document.getElementById("main");
const spinner = document.getElementById("spinner");
let sysInfoInterval;
let user = {};
let navbar;

document.addEventListener("DOMContentLoaded", async () => {
  // Try to get user if available (if not the error will redirect to #login)
  user = await Auth.GetUser();
  navbar = new Navbar(user);
  navbar.mount("navbar");
  window.addEventListener("hashchange", navigate);
  navigate();
});

async function navigate() {
  clearInterval(sysInfoInterval);
  switch (location.hash) {
    case "#apps":
      load(mountPoint, async function () {
        await Apps.mount("main", user);
      });
      break;
    case "#davs":
      load(mountPoint, async function () {
        await Davs.mount("main", user);
      });
      break;
    case "#users":
      load(mountPoint, async function () {
        await Users.mount("main", user);
      });
      break;
    case "#login":
      load(mountPoint, async function () {
        let login = new Login(user, navbar);
        login.mount("main");
      });
      break;
    case "#sysinfo":
      load(mountPoint, async function () {
        sysInfoInterval = await Sysinfo.mount("main", user);
      });
      break;
    default:
      location.hash = "#apps";
      break;
  }
  navbar.SetActiveItem();
}

async function load(element, domAlteration) {
  await AnimateCSS(element, "fadeOut");
  element.classList.add("is-hidden");
  // Start the alteration
  const alteration = domAlteration();
  spinner.classList.remove("is-hidden");
  await AnimateCSS(spinner, "fadeIn");
  await alteration; // Await for alteration end
  await AnimateCSS(spinner, "fadeOut");
  spinner.classList.add("is-hidden");
  element.classList.remove("is-hidden");
  AnimateCSS(element, "fadeIn");
}
