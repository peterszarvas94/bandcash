window.safeParseInt = function(value) {
  const parsed = parseInt(value, 10);
  return isNaN(parsed) ? 0 : parsed;
};

window.safeParseFloat = function(value) {
  const parsed = parseFloat(value);
  return isNaN(parsed) ? 0 : parsed;
};
