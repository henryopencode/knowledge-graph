'use strict';
const form = document.getElementById('login-form');
form.addEventListener('submit', async event => {
  event.preventDefault();
  const button = document.getElementById('login-submit');
  const message = document.getElementById('login-error');
  button.disabled = true; message.textContent = '';
  try {
    const response = await fetch(new URL('./api/login', location.href), {
      method: 'POST', headers: { 'Content-Type': 'application/json', 'X-DevGraph-Request': '1' },
      body: JSON.stringify({ username: document.getElementById('username').value.trim(), password: document.getElementById('password').value }),
    });
    const result = await response.json();
    if (!response.ok) throw new Error(result.error || '登录失败，请重试。');
    location.reload();
  } catch (error) { message.textContent = error.message; }
  finally { button.disabled = false; }
});
