# Constants
$LOGIN="user"
$PASSWORD="password"
$LOGIN_URL="https://vestibule.127.0.0.1.nip.io:1443/Login"
$FILES_URL= "https://admindav.vestibule.127.0.0.1.nip.io:1443"

# Mount
$body = "{`"login`":`"$LOGIN`",`"password`":`"$PASSWORD`"})"
Invoke-WebRequest -Uri $LOGIN_URL -Body $body -Method POST -SessionVariable websession -UseBasicParsing
$token = $websession.Cookies.GetCookies($LOGIN_URL)["auth_token"].value

$rename = New-Object -ComObject Shell.Application

net use T: $FILES_URL $token /user:$LOGIN /persistent:No
$rename.NameSpace("T:\").Self.Name = 'Test'  
