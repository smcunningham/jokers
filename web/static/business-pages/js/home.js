function randomJoke() {
    var e = document.getElementById("rJoke");
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            e.outerHTML = xhr.responseText;
        }
    }
    xhr.open("GET", "/jokes/random", true);
    try { xhr.send(); } catch (err) { console.log(err)}
}

function personalJoke() {
    var e = document.getElementById("pJoke");
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            e.outerHTML = xhr.responseText;
        }
    }
    xhr.open("GET", "/jokes/personal", true);
    try { xhr.send(); } catch (err) { console.log(err)}
}