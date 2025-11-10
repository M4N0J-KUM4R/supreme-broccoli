// Mobile menu toggle functionality
function toggleMobileMenu() {
  const navLinks = document.querySelector('.nav-links');
  const menuToggle = document.querySelector('.mobile-menu-toggle');
  
  if (navLinks && menuToggle) {
    navLinks.classList.toggle('mobile-active');
    menuToggle.classList.toggle('active');
    
    // Update aria-expanded attribute for accessibility
    const isExpanded = navLinks.classList.contains('mobile-active');
    menuToggle.setAttribute('aria-expanded', isExpanded);
  }
}

// Close mobile menu when clicking outside
document.addEventListener('click', function(event) {
  const nav = document.querySelector('.main-nav');
  const navLinks = document.querySelector('.nav-links');
  const menuToggle = document.querySelector('.mobile-menu-toggle');
  
  if (nav && navLinks && menuToggle && 
      !nav.contains(event.target) && 
      navLinks.classList.contains('mobile-active')) {
    navLinks.classList.remove('mobile-active');
    menuToggle.classList.remove('active');
    menuToggle.setAttribute('aria-expanded', 'false');
  }
});
