// Contact form validation and interactivity

const contactForm = document.getElementById('contactForm');
const nameInput = document.getElementById('name');
const emailInput = document.getElementById('email');
const subjectInput = document.getElementById('subject');
const messageInput = document.getElementById('message');

// Error display elements
const nameError = document.getElementById('nameError');
const emailError = document.getElementById('emailError');
const subjectError = document.getElementById('subjectError');
const messageError = document.getElementById('messageError');

// Validation functions
function validateName() {
  const name = nameInput.value.trim();
  if (name === '') {
    showError(nameError, 'Name is required');
    return false;
  }
  if (name.length < 2) {
    showError(nameError, 'Name must be at least 2 characters');
    return false;
  }
  clearError(nameError);
  return true;
}

function validateEmail() {
  const email = emailInput.value.trim();
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  
  if (email === '') {
    showError(emailError, 'Email is required');
    return false;
  }
  if (!emailRegex.test(email)) {
    showError(emailError, 'Please enter a valid email address');
    return false;
  }
  clearError(emailError);
  return true;
}

function validateSubject() {
  const subject = subjectInput.value.trim();
  if (subject === '') {
    showError(subjectError, 'Subject is required');
    return false;
  }
  if (subject.length < 3) {
    showError(subjectError, 'Subject must be at least 3 characters');
    return false;
  }
  clearError(subjectError);
  return true;
}

function validateMessage() {
  const message = messageInput.value.trim();
  if (message === '') {
    showError(messageError, 'Message is required');
    return false;
  }
  if (message.length < 10) {
    showError(messageError, 'Message must be at least 10 characters');
    return false;
  }
  clearError(messageError);
  return true;
}

function showError(element, message) {
  element.textContent = message;
  element.style.display = 'block';
  element.parentElement.querySelector('.form-input, .form-textarea').classList.add('error');
}

function clearError(element) {
  element.textContent = '';
  element.style.display = 'none';
  element.parentElement.querySelector('.form-input, .form-textarea').classList.remove('error');
}

// Real-time validation on blur
if (nameInput) nameInput.addEventListener('blur', validateName);
if (emailInput) emailInput.addEventListener('blur', validateEmail);
if (subjectInput) subjectInput.addEventListener('blur', validateSubject);
if (messageInput) messageInput.addEventListener('blur', validateMessage);

// Form submission validation
if (contactForm) {
  contactForm.addEventListener('submit', function(e) {
    const isNameValid = validateName();
    const isEmailValid = validateEmail();
    const isSubjectValid = validateSubject();
    const isMessageValid = validateMessage();

    if (!isNameValid || !isEmailValid || !isSubjectValid || !isMessageValid) {
      e.preventDefault();
    }
  });
}
