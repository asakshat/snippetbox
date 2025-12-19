
var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
	var link = navLinks[i]
	if (link.getAttribute('href') == window.location.pathname) {
		link.classList.add("live");
		break;
	}
}

document.addEventListener('DOMContentLoaded', function () {
	const flash = document.getElementById('flash-message');
	if (flash) {
		setTimeout(() => {
			flash.style.transition = 'opacity 0.5s';
			flash.style.opacity = '0';
			setTimeout(() => flash.remove(), 500);
		}, 3000);
	}
});

document.body.addEventListener('htmx:afterSwap', function () {
	const flash = document.getElementById('flash-message');
	if (flash) {
		setTimeout(() => {
			flash.style.transition = 'opacity 0.5s';
			flash.style.opacity = '0';
			setTimeout(() => flash.remove(), 500);
		}, 3000);
	}
});