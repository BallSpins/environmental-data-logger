#pragma once
#include <Arduino.h>

class EMAFilter {
  private:
    float alpha;
    float currentEMA;
    bool isInitialized;
  public:
    EMAFilter(float alphaValue);
    float update(float newValue);
    // reset();
    // setAlpha(float alphaValue);
};
