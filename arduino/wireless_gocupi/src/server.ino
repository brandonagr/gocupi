/**
Test out using NRF2401L+ for wireless communication
 */

#include <SPI.h>
#include <Mirf.h>
#include <nRF24L01.h>
#include <MirfHardwareSpiDriver.h>

const int MIRF_CE_PIN = A5;
const int MIRF_CSN_PIN = A6;

void setup(){

 // Setup pins / SPI.
  Mirf.cePin = MIRF_CE_PIN;
  Mirf.csnPin = MIRF_CSN_PIN;
  Mirf.spi = &MirfHardwareSpi;
  Mirf.init();
  
  Mirf.payload = 1;
  Mirf.channel = 1;
  Mirf.setTADDR((byte *)"clie1");
}

void loop(){

  byte data = 1;
  Mirf.send((byte*)&data);
  //while (Mirf.isSending()) {
  //}
  delay(1000);
  data = 0;
  Mirf.send((byte*)&data);
  //while (Mirf.isSending()) {
  //}
  delay(1000);
}
