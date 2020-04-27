import * as Apps from "/components/apps/apps.js";
import * as Davs from "/components/davs/davs.js";
import * as Users from "/components/users/users.js";
import * as Login from "/components/login/login.js";
import * as Sysinfo from "/components/sysinfo/sysinfo.js";
import * as Navbar from "/components/navbar/navbar.js";
import { AnimateCSS } from "/services/common/common.js";

const mountPoint = document.getElementById("main");
const spinner = document.getElementById("spinner");
let sysInfoInterval;

document.addEventListener("DOMContentLoaded", function () {
  Navbar.mount("navbar");
  window.addEventListener("hashchange", navigate);
  navigate();
});

async function navigate() {
  clearInterval(sysInfoInterval);
  switch (location.hash) {
    case "#apps":
      load(mountPoint, async function () {
        await Apps.mount("main");
      });
      break;
    case "#davs":
      load(mountPoint, async function () {
        await Davs.mount("main");
      });
      break;
    case "#users":
      load(mountPoint, async function () {
        await Users.mount("main");
      });
      break;
    case "#login":
      load(mountPoint, async function () {
        await Login.mount("main");
      });
      break;
    case "#sysinfo":
      load(mountPoint, async function () {
        sysInfoInterval = await Sysinfo.mount("main");
      });
      break;
    default:
      location.hash = "#apps";
      break;
  }
  Navbar.SetActiveItem();
}

async function load(element, domAlteration) {
  AnimateCSS(element, "fadeOut", async function () {
    element.classList.add("is-hidden");
    spinner.classList.remove("is-hidden");
    AnimateCSS(spinner, "fadeIn");
    if (typeof domAlteration === "function") {
      await domAlteration();
      AnimateCSS(spinner, "fadeOut", function () {
        spinner.classList.add("is-hidden");
      });
      element.classList.remove("is-hidden");
      AnimateCSS(element, "fadeIn");
    }
  });
}
