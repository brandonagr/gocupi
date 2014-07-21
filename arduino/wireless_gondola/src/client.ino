/**
Test out using NRF2401L+ for wireless communication
**/

#include <SPI.h>
#include <Mirf.h>
#include <nRF24L01.h>
#include <MirfHardwareSpiDriver.h>
#include <Servo.h>

Servo penUpServo;
const int PENUP_SERVO_PIN = 9;
// this is SPI SCK pin, so cant use it as status led
//const int STATUS_LED_PIN = 13;

const int MIRF_CE_PIN = A5;
const int MIRF_CSN_PIN = A6;

byte readDataFromMirf = 0;
long blinkDelayMillis = 250;
byte currentLedStatus = 0;
unsigned long previousTime = 0;

void setup(){
  // setup servo
  penUpServo.attach(PENUP_SERVO_PIN);
  //pinMode(STATUS_LED_PIN, OUTPUT);
  //digitalWrite(STATUS_LED_PIN, 0);

  // Setup pins / SPI.
  Mirf.cePin = MIRF_CE_PIN;
  Mirf.csnPin = MIRF_CSN_PIN;
  Mirf.spi = &MirfHardwareSpi;
  Mirf.init();
  
  // Configure reciving address.
  Mirf.setRADDR((byte *)"clie1");
  Mirf.payload = sizeof(byte);
  Mirf.channel = 1;
  Mirf.config();
}

void loop(){
  // update status led
  // when sketch first start led blinks rapidly, until after it receives at least one byte over
  unsigned long curTime = millis();
  blinkDelayMillis -= (curTime - previousTime);
  if (blinkDelayMillis < 0) {
    currentLedStatus = !currentLedStatus;
    //digitalWrite(STATUS_LED_PIN, currentLedStatus);

    if (readDataFromMirf) {
      if (currentLedStatus) {
// += would be more accurate, but want to prevent propogation of potential errors like time overflow
        blinkDelayMillis = 500;
      } else {
        blinkDelayMillis = 2000;
      }
    } else {
      blinkDelayMillis = 1000;
    }
  }
  previousTime = curTime;


  byte receiveData;

  if (Mirf.dataReady()) {
    readDataFromMirf = 1;
    Mirf.getData((byte *) &receiveData);

    if (receiveData) {
      penUpServo.write(150);
    } else {
      penUpServo.write(0);
    }
  }
} 
