#include <Filter.h>

EMAFilter::EMAFilter(float alphaValue) {
  if (alphaValue <= 0.0f) alpha = 0.01f;
  else if (alphaValue > 1.0f) alpha = 1.0f;
  else alpha = alphaValue;
  
  currentEMA = 0.0f;
  isInitialized = false;
}

float EMAFilter::update(float newValue) {
  if (!isInitialized) {
    currentEMA = newValue;
    isInitialized = true;
  } else {
    currentEMA = (alpha * newValue) + ((1.0f - alpha) * currentEMA);
  }

  return currentEMA;
}
