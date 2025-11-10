// Settings page interactivity

// Font size slider preview
const fontSizeSlider = document.getElementById('terminalFontSize');
const fontSizeValue = document.getElementById('fontSizeValue');

if (fontSizeSlider && fontSizeValue) {
  fontSizeSlider.addEventListener('input', function() {
    fontSizeValue.textContent = this.value + 'px';
  });
}

// Color scheme preview
const colorSchemeSelect = document.getElementById('terminalColorScheme');
const colorSchemePreview = document.getElementById('colorSchemePreview');

const colorSchemes = {
  dark: { bg: '#1e1e1e', fg: '#d4d4d4' },
  light: { bg: '#ffffff', fg: '#333333' },
  solarized: { bg: '#002b36', fg: '#839496' },
  monokai: { bg: '#272822', fg: '#f8f8f2' }
};

function updateColorSchemePreview() {
  if (!colorSchemeSelect || !colorSchemePreview) return;
  
  const scheme = colorSchemes[colorSchemeSelect.value];
  if (scheme) {
    colorSchemePreview.style.backgroundColor = scheme.bg;
    colorSchemePreview.style.color = scheme.fg;
    colorSchemePreview.textContent = `$ echo "Preview: ${colorSchemeSelect.value}"`;
  }
}

if (colorSchemeSelect) {
  colorSchemeSelect.addEventListener('change', updateColorSchemePreview);
  updateColorSchemePreview(); // Initial preview
}

// Client-side form validation
const settingsForm = document.getElementById('settingsForm');

if (settingsForm) {
  settingsForm.addEventListener('submit', function(e) {
    let isValid = true;
    const errors = [];

    // Validate font size
    const fontSize = parseInt(fontSizeSlider.value);
    if (fontSize < 10 || fontSize > 24) {
      isValid = false;
      errors.push('Font size must be between 10 and 24 pixels');
    }

    // Validate color scheme
    const colorScheme = colorSchemeSelect.value;
    if (!['dark', 'light', 'solarized', 'monokai'].includes(colorScheme)) {
      isValid = false;
      errors.push('Please select a valid color scheme');
    }

    // Validate cursor style
    const cursorStyle = document.getElementById('terminalCursorStyle').value;
    if (!['block', 'underline', 'bar'].includes(cursorStyle)) {
      isValid = false;
      errors.push('Please select a valid cursor style');
    }

    if (!isValid) {
      e.preventDefault();
      alert('Please fix the following errors:\n\n' + errors.join('\n'));
    }
  });
}
