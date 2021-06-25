// Imports
import { HandleError } from "/services/common/errors.js";
import { IsEmpty } from "/services/common/common.js";

// Local variables
let user = {};

export async function GetUser() {
  if (!IsEmpty(user)) {
    return user;
  }
  try {
    const response = await fetch("/api/common/WhoAmI");
    if (response.status !== 200) {
      throw new Error(`Not authenticated (status ${response.status})`);
    }
    Object.assign(user, await response.json());
    // Redirect to original subdomain if login was displayed after an authentication error on the original subdomain
    try {
      const redirectAfterLogin = document.cookie
        .split("; ")
        .find((row) => row.startsWith("redirectAfterLogin="))
        .split(/=(.+)/)[1];

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
  Object.keys(user).forEach((key) => {
    delete user[key];
  });
}
