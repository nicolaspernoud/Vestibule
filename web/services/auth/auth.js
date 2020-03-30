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
  } catch (e) {
    HandleError(e);
  }
  return user;
}

export function DeleteUser() {
  user = undefined;
}
