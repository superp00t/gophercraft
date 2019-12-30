window.getAuthToken = () => {
    var data = window.localStorage["gcraft-token"];
    if (data == null)
        return "";

    return data;
}

window.getAuthStatus = async () => {
    var data = await fetch("/v1/getAuthStatus", {
        method: "POST",
        body:   JSON.stringify({
            token: getAuthToken()
        })
    });

    var response = await data.json();

    return response;
}

window.addAccountButton = (element, name, url) => {
    var btn = document.createElement("a");
    btn.classList.add("accountButton");
    btn.setAttribute("href", url);

    var label = document.createElement("span");
    label.innerText = name;
    btn.appendChild(label)
    
    element.appendChild(btn);
}

window.makeTextInput = (type, placeholder, className, id) => {
    var input = document.createElement("input");
    input.setAttribute("type", type);
    input.setAttribute("placeholder", placeholder);
    if (id)
        input.setAttribute("id", id);
    input.className = className;
    return input;
}

window.addBreak = (el) => {
    var br = document.createElement("br");
    el.appendChild(br);
}

window.addLoginForm = (el) => {
    addBreak(el);

    var signUpLink = document.createElement("a");
    signUpLink.setAttribute("href", "/signUp");
    signUpLink.innerText = "Sign up";
    el.appendChild(signUpLink);
    addBreak(el);

    var username = makeTextInput("text", "username", "textInput", "signInUsername");
    el.appendChild(username);
    addBreak(el);

    var pwd = makeTextInput("password", "password", "textInput", "signInPassword");
    el.appendChild(pwd);
    addBreak(el);

    var login = document.createElement("button");
    login.innerText = "Sign in";
    login.addEventListener("click", async () => {
        var response = await fetch("/v1/signIn", {
            method: "POST",
            body: JSON.stringify({
                username: document.querySelector("#signInUsername").value,
                password: document.querySelector("#signInPassword").value,
           })
        });

        var data = await response.json();

        if (data.error !== "") {
            alert(data.error);
        } else {
            window.localStorage["gcraft-token"] = data.webToken;
            window.resetAccount();
        }
    });

    el.appendChild(login);
}

window.resetAccount = () => {
    window.getAuthStatus()
    .then((data) => {
        var area = document.querySelector(".accountInfoArea");
        area.innerHTML = "";
        var status = document.createElement("span");
        status.classList.add("accountLoginStatus");

        if (data.valid) {
            var statusIco = document.createElement("img");
            statusIco.setAttribute("src", "/assets/ok_icon.png");
            status.innerText = "Hello, " + data.account + "!";
            status.classList.add("ok");
            area.appendChild(statusIco);
            area.appendChild(status);
        } else {
            area.appendChild(status);
            status.innerText = "You are not logged in.";
            window.addLoginForm(area);
        }
     });
}