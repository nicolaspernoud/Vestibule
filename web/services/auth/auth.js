// Imports
import { HandleError } from "/services/common/errors.js";

// Local variables
let user;

export async function GetUser() {
  if (user != undefined) {
    return user;
  }
  try {
    const response = await fetch("/api/common/WhoAmI");
    if (response.status !== 200) {
      throw new Error(`Not authenticated (status ${response.status})`);
    }
    user = await response.json();
    // Redirect to original subdomain if login was displayed after an authentication error on the original subdomain
    try {
      const redirectAfterLogin = document.cookie
        .split("; ")
        .find((row) => row.startsWith("redirectAfterLogin="))
        .split("=")[1];

      if (redirectAfterLogin != "" && redirectAfterLogin != null) {
        window.location.replace("https://" + redirectAfterLogin);
      }
    } catch (e) {}
  } catch (e) {
    HandleError(e);
  }
  return user;
}

export function DeleteUser() {
  user = undefined;
}
