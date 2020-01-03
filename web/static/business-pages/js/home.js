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

function customJoke() {
    console.log("inside customJoke()");

    $("#customModal").modal("hide");
}

$('#customModal').on('hide.bs.modal', function () {
    var person = {first: $("#firstname").val(), last: $("#lastname").val()};
    console.log(person)
    
    data = JSON.stringify(person);
    e = document.getElementById("cJoke");

    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            e.outerHTML = xhr.responseText;
        }
    }
    xhr.open("POST", "/jokes/custom", true);
    xhr.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    try { xhr.send(data); } catch (err) { console.log(err)}
})
