// Imports
import { RandomString } from "/services/common/common.js";
import { HandleError } from "/services/common/errors.js";

export async function mount(where, user) {
  const infoComponent = new Sysinfo(user);
  await infoComponent.mount(where);
  return setInterval(() => {
    infoComponent.update();
  }, 1000);
}

class Sysinfo {
  constructor(user) {
    this.user = user;
    // Random id seed
    this.prefix = RandomString(8);
  }

  async mount(mountpoint) {
    const mnt = document.getElementById(mountpoint);
    mnt.innerHTML = /* HTML */ `
      <div class="card">
        <div id="${this.prefix}-sysinfo" class="card-content"></div>
      </div>
    `;
    await this.update();
  }

  async update() {
    const content = document.getElementById(`${this.prefix}-sysinfo`);
    try {
      const response = await fetch("/api/admin/sysinfo/", {
        method: "GET",
        headers: new Headers({
          "XSRF-Token": this.user.xsrftoken,
        }),
        credentials: "include",
      });
      if (response.status !== 200) {
        throw new Error(`System information could not be fetched (status ${response.status})`);
      }
      const info = await response.json();
      content.innerHTML = this.computeTemplate(info);
    } catch (e) {
      HandleError(e);
    }
  }

  computeTemplate(info) {
    const mu = (info.totalram - info.freeram) / info.totalram;
    const du = info.usedgb / info.totalgb;
    const dfree = info.totalgb - info.usedgb;
    return /* HTML */ `
      <div class="content">
        ${info.load !== undefined
          ? /* HTML */ `
              <h1>CPU usage</h1>
              <p>
                <progress class="progress small-radius is-small is-${GetColor(info.load)}" value="${info.load}" max="1"></progress>
                ${(info.load * 100).toFixed(0)} %
              </p>
            `
          : ""}
        ${info.freeram !== undefined
          ? /* HTML */ `
              <h1>Memory usage</h1>
              <p>
                <progress class="progress small-radius is-small is-${GetColor(mu)}" value="${info.totalram - info.freeram}" max="${info.totalram}"></progress>
                ${(info.freeram / Math.pow(2, 20)).toFixed(2)} GB free
              </p>
            `
          : ""}
        ${info.usedgb !== undefined
          ? /* HTML */ `
              <h1>Disk usage</h1>
              <p>
                <progress class="progress small-radius is-small is-${GetColor(du)}" value="${info.usedgb}" max="${info.totalgb}"></progress>
                ${dfree} GB free
              </p>
            `
          : ""}
        ${info.uptime !== undefined
          ? /* HTML */ `
              <h1>Uptime</h1>
              <p>${secondsToDhms(info.uptime / Math.pow(10, 9))}</p>
            `
          : ""}
      </div>
    `;
  }
}

function secondsToDhms(seconds) {
  seconds = Number(seconds);
  const d = Math.floor(seconds / (3600 * 24));
  const h = Math.floor((seconds % (3600 * 24)) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);

  const dDisplay = d > 0 ? d + (d == 1 ? " day, " : " days, ") : "";
  const hDisplay = h > 0 ? h + (h == 1 ? " hour, " : " hours, ") : "";
  const mDisplay = m > 0 ? m + (m == 1 ? " minute, " : " minutes, ") : "";
  const sDisplay = s > 0 ? s + (s == 1 ? " second" : " seconds") : "0 second";
  return dDisplay + hDisplay + mDisplay + sDisplay;
}

export function GetColor(load) {
  switch (true) {
    case load >= 0.9:
      return "danger";
    case load >= 0.75 && load < 0.9:
      return "warning";
    default:
      return "success";
  }
}
