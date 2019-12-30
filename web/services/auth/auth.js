// Imports
import * as Messages from "/services/messages/messages.js";
import * as Login from "/components/login/login.js";

// Local variables
let user;

export async function GetUser() {
  if (user != undefined) {
    return user;
  }
  try {
    const response = await fetch("/api/common/WhoAmI");
    if (response.status !== 200) {
      throw new Error(`The user is unknown (status ${response.status})`);
    }
    user = await response.json();
    return user;
  } catch (e) {
    Messages.Show("is-warning", e);
    console.error(e);
    location.hash = "#login";
  }
}
