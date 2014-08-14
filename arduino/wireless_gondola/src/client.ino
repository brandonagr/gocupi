/**
Test out using NRF2401L+ for wireless communication
**/

#include <SPI.h>
#include <nRF24L01.h>
#include <RF24.h>
#include <RF24_config.h>
#include <Servo.h>

Servo penUpServo;
const int PENUP_SERVO_PIN = 9;
// this is SPI SCK pin, so cant use it as status led
const int STATUS_LED_PIN = 8;

const int CE_PIN = A5;
const int CSN_PIN = A6;

RF24 radio(CE_PIN, CSN_PIN);

byte readDataFromMirf = 0;
long blinkDelayMillis = 250;
bool currentLedStatus = false;
unsigned long previousTime = 0;

void setup(){
  
//  Serial.begin(57600);
//  Serial.println("Setup");
  
  // setup servo
  penUpServo.attach(PENUP_SERVO_PIN);
  pinMode(STATUS_LED_PIN, OUTPUT);
  digitalWrite(STATUS_LED_PIN, 0);
  delay(1000);
  digitalWrite(STATUS_LED_PIN, 1);
  delay(1000);
  digitalWrite(STATUS_LED_PIN, 0);
  delay(1000);

  radio.begin();
  radio.setRetries(15, 15); // 2*250=500us between tries, 5 tries
  radio.setPayloadSize(4);
  radio.openReadingPipe(1, 0xF0F0F0F0E1LL);
  radio.openWritingPipe(0xF0F0F0F0D2LL);
  radio.startListening();
  
//  Serial.println("Setup done");
}

void loop(){
  // update status led
  // when sketch first start led blinks rapidly, until after it receives at least one byte over
  unsigned long curTime = millis();
  blinkDelayMillis -= (curTime - previousTime);
  if (blinkDelayMillis < 0) {
    currentLedStatus = !currentLedStatus;
    digitalWrite(STATUS_LED_PIN, currentLedStatus);
    
    if (currentLedStatus) {
      blinkDelayMillis = 250;
    } else {
      blinkDelayMillis = 750;
    }
/*
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
*/
  }
  previousTime = curTime;

  radio.startListening();
  delay(100);

  long receiveData;
  if (radio.available()) {
    
    readDataFromMirf = 1;
    bool result = radio.read(&receiveData, sizeof(long));
    
    if (result) {
//      Serial.print("Got data: ");
//      Serial.println(receiveData);

      if (receiveData) {
        penUpServo.write(150);

        currentLedStatus = 1;
        digitalWrite(STATUS_LED_PIN, 1);
        blinkDelayMillis = 1000;
      } else {
        penUpServo.write(0);

        currentLedStatus = 1;
        digitalWrite(STATUS_LED_PIN, 1);
        blinkDelayMillis = 2000;
      }
    } else {
//      Serial.println("Failed to read");
        digitalWrite(STATUS_LED_PIN, 0);
    }
  }
}

