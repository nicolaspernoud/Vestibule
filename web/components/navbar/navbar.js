// Imports
import * as brand from "/assets/brand/brand.js";
import { AnimateCSS, IsEmpty } from "/services/common/common.js";

export class Navbar {
  constructor(user) {
    this.user = user;
  }

  // local variables
  user;
  menu;

  mount(mountpoint) {
    const where = document.getElementById(mountpoint);
    window.document.title = brand.windowTitle;
    where.innerHTML = /* HTML */ `
      <div class="navbar-brand">
        <a class="navbar-item is-size-4" href="/"><img src="assets/brand/logo.svg" alt="logo" />${brand.navTitle}</a>
        <a role="button" id="navbar-burger" class="navbar-burger burger" aria-label="menu" aria-expanded="false">
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
        </a>
      </div>
      <div id="navbar-menu" class="navbar-menu"></div>
    `;
    // Hamburger menu
    const burger = document.getElementById("navbar-burger");
    this.menu = document.getElementById("navbar-menu");
    const openClose = async (e) => {
      if (burger.classList.contains("is-active")) {
        await AnimateCSS(this.menu, "slideOutRight");
        this.menu.classList.remove("is-active");
        burger.classList.remove("is-active");
      } else {
        if (e.srcElement == burger || e.srcElement.offsetParent == burger) {
          this.menu.classList.add("is-active");
          burger.classList.add("is-active");
          AnimateCSS(this.menu, "slideInRight");
        }
      }
    };
    burger.addEventListener("click", openClose);
    this.menu.addEventListener("click", openClose);
    this.CreateMenu();
  }

  async CreateMenu() {
    this.menu.innerHTML = /* HTML */ `
      <div class="navbar-start">
        ${IsEmpty(this.user)
          ? ``
          : /* HTML */ `
              <a id="navbar-apps" class="navbar-item" href="#apps"><i class="navbar-menu-icon fas fa-home"></i>Apps</a>
              <a id="navbar-davs" class="navbar-item" href="#davs"><i class="navbar-menu-icon fas fa-folder-open"></i>Files</a>
              ${this.user.isAdmin
                ? /* HTML */ `
                    <a id="navbar-users" class="navbar-item" href="#users"><i class="navbar-menu-icon fas fa-users"></i>Users</a>
                    <a id="navbar-sysinfo" class="navbar-item" href="#sysinfo"><i class="navbar-menu-icon fas fa-stethoscope"></i>System information</a>
                  `
                : ""}
            `}
      </div>
      <div class="navbar-end">
        ${IsEmpty(this.user)
          ? /* HTML */ ` <a class="navbar-item" href="#login"><i class="navbar-menu-icon fas fa-sign-in-alt"></i>Log in</a> `
          : /* HTML */ ` <a class="navbar-item" href="/Logout"><i class="navbar-menu-icon fas fa-sign-out-alt"></i>Log out</a> `}
      </div>
    `;
    this.SetActiveItem();
  }

  SetActiveItem() {
    const items = document.getElementsByClassName("navbar-item");
    for (const i of items) {
      i.classList.remove("is-active");
      if (i.id == "navbar-" + location.hash.substring(1)) {
        i.classList.add("is-active");
      }
    }
  }
}
