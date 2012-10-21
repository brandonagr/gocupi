/*
  Run a stepper driver, by reading step data over serial
 
 This example code is in the public domain.
 */

int ledPins[5] = {
  2,3,4,5,8}; // the pins of all of the leds, first 4 are status lights, 5th is error indicator
int leftStepPin = 7;
int leftDirPin = 6;
int rightStepPin = 9;
int rightDirPin = 10;

// Global variables
// --------------------------------------
const unsigned int TIME_SLICE_US = 2000; // number of microseconds per time step

const unsigned int MOVE_DATA_CAPACITY = 1024;
byte moveData[MOVE_DATA_CAPACITY]; // buffer of move data, circular buffer
unsigned int moveDataStart = 0; // where data is currently being read from
unsigned int moveDataLength = 0; // the number of items in the moveDataBuffer
unsigned int moveDataRequestPending = 0; // number of bytes requested

unsigned int currentTimeSlice = 0; // the amount of microseconds that have been observed
unsigned int timeSinceStepLeft = 0;
unsigned int timeSinceStepRight = 0;
unsigned int stepDelayLeft = 0;
unsigned int stepDelayRight = 0;
unsigned int stepsLeft = 0;
unsigned int stepsRight = 0;
unsigned int targetStepsLeft = 0;
unsigned int targetStepsRight = 0;
byte leftDir = 0;
byte rightDir = 0;

unsigned long currentTime = 0; // microseconds
unsigned long previousTime = 0; // microseconds since last
unsigned int deltaTime = 0; // microseconds since last loop


// setup
// --------------------------------------
void setup() {
  Serial.begin(57600);
  Serial.setTimeout(1);

  // setup pins
  for(int ledIndex = 0; ledIndex < 5; ledIndex++) {
    pinMode(ledPins[ledIndex], OUTPUT);
    digitalWrite(ledPins[ledIndex], HIGH);
  }	
  pinMode(leftStepPin, OUTPUT);
  pinMode(leftDirPin, OUTPUT);
  pinMode(rightStepPin, OUTPUT);
  pinMode(rightDirPin, OUTPUT);	

  SetStepVariables();

  delay(500);
  UpdateErrorLed(false);

}

// Main execution loop
// --------------------------------------
void loop() {

  currentTime = micros();
  deltaTime = (int)(currentTime - previousTime); // assuming we wont go over 65,536 microseconds per loop

  UpdateStepperPins();
  ReadSerialMoveData();
  RequestMoreSerialMoveData();

  previousTime = currentTime;
}

// Update status leds
// --------------------------------------
void UpdateStatusLeds(int value) {

  // output the time to the leds in binary
  digitalWrite(ledPins[0], value & 0x1);
  digitalWrite(ledPins[1], value & 0x2);
  digitalWrite(ledPins[2], value & 0x4);
  digitalWrite(ledPins[3], value & 0x8);
}

// Update status leds
// --------------------------------------
void UpdateErrorLed(boolean value) {
  digitalWrite(ledPins[4], value);
}

// Step
// --------------------------------------
void Step(int stepPin, int dirPin, boolean dir) {

  digitalWrite(dirPin, dir);

  digitalWrite(stepPin, LOW);
  digitalWrite(stepPin, HIGH);
}

// Update stepper pins
// --------------------------------------
void UpdateStepperPins() {

  currentTimeSlice += deltaTime;

  // need to make sure the correct number of steps are taken
  // want to be able to

  timeSinceStepLeft += deltaTime;
  while (timeSinceStepLeft >= stepDelayLeft && stepsLeft < targetStepsLeft) {
    Step(leftStepPin, leftDirPin, leftDir);
    stepsLeft++;
    timeSinceStepLeft -= stepDelayLeft;
  }

  timeSinceStepRight += deltaTime;
  while (timeSinceStepRight >= stepDelayRight && stepsRight < targetStepsRight) {
    Step(rightStepPin, rightDirPin, rightDir);
    stepsRight++;
    timeSinceStepRight -= stepDelayRight;
  }

  // move to next time slice
  if (currentTimeSlice >= TIME_SLICE_US) {

    currentTimeSlice -= TIME_SLICE_US;
    SetStepVariables();
    
    // how could sampling be so slow that we jumped past an entire time slice !?
    UpdateErrorLed(currentTimeSlice >= TIME_SLICE_US);
  }
}

// Set all step variables based on the data currently in the buffer
// --------------------------------------
void SetStepVariables() {

  stepsLeft = 0;
  stepsRight = 0;
  timeSinceStepLeft = 0;
  timeSinceStepRight = 0;

  // just delay and do nothing if there is no data
  if (moveDataLength < 2) {
    stepDelayLeft = 0;
    stepDelayRight = 0;
    targetStepsLeft = 0;
    targetStepsRight = 0;
    return;
  }

  byte data = MoveDataGet();
  leftDir = data & 0x80;
  targetStepsLeft = data & 0x7F;
  stepDelayLeft = (TIME_SLICE_US - currentTimeSlice) / targetStepsLeft;

  data = MoveDataGet();
  rightDir = data & 0x80;
  targetStepsRight = data & 0x7F;
  stepDelayRight = (TIME_SLICE_US - currentTimeSlice) / targetStepsRight;
}


// Read serial data if its available
// --------------------------------------
void ReadSerialMoveData() {

  while(Serial.available()) {
    MoveDataPut(Serial.read());
    moveDataRequestPending--;
    UpdateStatusLeds(moveDataRequestPending);
  }
}

// Put a value onto the end of the move data buffer
// --------------------------------------
void MoveDataPut(byte value) {

  int writePosition = moveDataStart + moveDataLength;
  if (writePosition >= MOVE_DATA_CAPACITY) {
    writePosition = writePosition - MOVE_DATA_CAPACITY;
  }

  moveData[writePosition] = value;

  if (moveDataLength == MOVE_DATA_CAPACITY) { // full, overwrite existing data
    UpdateErrorLed(true);
    moveDataStart++;
    if (moveDataStart == MOVE_DATA_CAPACITY) {
      moveDataStart = 0;
    }
  } 
  else {
    moveDataLength++;
  }
}

// Return the amount of data sitting in the moveData buffer
// --------------------------------------
byte MoveDataGet() {

  if (moveDataLength == 0) {
    return 0;
  }

  byte result = moveData[moveDataStart];
  moveDataStart++;
  if (moveDataStart == MOVE_DATA_CAPACITY) {
    moveDataStart = 0;
  }
  moveDataLength--;

  //UpdateStatusLeds(moveDataLength >> 6);

  return result;
}

// Return the amount of data sitting in the moveData buffer
// --------------------------------------
void RequestMoreSerialMoveData() {
  if (moveDataRequestPending > 0 || MOVE_DATA_CAPACITY - moveDataLength < 128)
    return;

  // request 128 bytes of data
  Serial.write(128);
  moveDataRequestPending = 128;
}


