/*
  Run a stepper driver, by reading step data over serial
 
 This example code is in the public domain.
 */

const int ledPins[5] = {
  2,3,4,5,8}; // the pins of all of the leds, first 4 are status lights, 5th is receive indicator
const int leftStepPin = 7;
const int leftDirPin = 6;
const int rightStepPin = 9;
const int rightDirPin = 10;

// Global variables
// --------------------------------------
const unsigned int TIME_SLICE_US = 2048; // number of microseconds per time step
const unsigned int TIME_SLICE_US_LOG = 11; // log base 2 of TIME_SLICE_US
const unsigned int POS_FACTOR = 32; // factor each position is multiplied by
const unsigned int POS_FACTOR_LOG = 5; // log base 2 of POS_FACTOR

const unsigned int MOVE_DATA_CAPACITY = 1024;
char moveData[MOVE_DATA_CAPACITY]; // buffer of move data, circular buffer
unsigned int moveDataStart = 0; // where data is currently being read from
unsigned int moveDataLength = 0; // the number of items in the moveDataBuffer
unsigned int moveDataRequestPending = 0; // number of bytes requested

char leftDelta, rightDelta; // delta in the current slice
long leftStartPos, rightStartPos; // start position for this slice
long leftCurPos, rightCurPos; // current position of the spools

unsigned long curTime; // current time in microseconds
unsigned long sliceStartTime; // start of current slice in microseconds


// setup
// --------------------------------------
void setup() {
  Serial.begin(57600);
  Serial.setTimeout(0);

  // setup pins
  for(int ledIndex = 0; ledIndex < 5; ledIndex++) {
    pinMode(ledPins[ledIndex], OUTPUT);
    digitalWrite(ledPins[ledIndex], HIGH);
  }	
  pinMode(leftStepPin, OUTPUT);
  pinMode(leftDirPin, OUTPUT);
  pinMode(rightStepPin, OUTPUT);
  pinMode(rightDirPin, OUTPUT);	

  ResetMovementVariables();

  delay(500);
  UpdateReceiveLed(false);
  UpdateStatusLeds(0);
}

// Reset all movement variables
// --------------------------------------
void ResetMovementVariables()
{
  leftDelta = rightDelta = leftStartPos = rightStartPos = leftCurPos = rightCurPos = 0;
  sliceStartTime = curTime;
}

// Main execution loop
// --------------------------------------
void loop() {

  curTime = micros();
  if (curTime < sliceStartTime) { // protect against 70 minute overflow
    sliceStartTime = 0;
  }

  // move to next slice if necessary
  long curSliceTime = curTime - sliceStartTime;
  while(curSliceTime > TIME_SLICE_US) {
    SetSliceVariables();
    curSliceTime -= TIME_SLICE_US;
    sliceStartTime += TIME_SLICE_US;
  }

  UpdateStepperPins(curSliceTime);
  ReadSerialMoveData();
  RequestMoreSerialMoveData();
}

// Update stepper pins
// --------------------------------------
void UpdateStepperPins(long curSliceTime) {

  long leftTarget = ((long(leftDelta) * curSliceTime) >> TIME_SLICE_US_LOG) + leftStartPos;
  long rightTarget = ((long(rightDelta) * curSliceTime) >> TIME_SLICE_US_LOG) + rightStartPos;

  int leftSteps = (leftTarget - leftCurPos) >> POS_FACTOR_LOG;
  int rightSteps = (rightTarget - rightCurPos) >> POS_FACTOR_LOG;

  boolean leftPositiveDir = true;
  if (leftSteps < 0) {
    leftPositiveDir = false;
    leftSteps = -leftSteps;
  }
  boolean rightPositiveDir = true;
  if (rightSteps < 0) {
    rightPositiveDir = false;
    rightSteps = -rightSteps;
  }

  do {
    if (leftSteps) {
      Step(leftStepPin, leftDirPin, leftPositiveDir);
      if (leftPositiveDir) {
        leftCurPos += POS_FACTOR;
      } else {
        leftCurPos -= POS_FACTOR;
      }
      leftSteps--;
      
      UpdateStatusLeds(leftCurPos >> 13);
    }

    if (rightSteps) {
      Step(rightStepPin, rightDirPin, rightPositiveDir);
      if (rightPositiveDir) {
        rightCurPos += POS_FACTOR;
      } else {
        rightCurPos -= POS_FACTOR;
      }
      rightSteps--;
    }

    if (leftSteps || rightSteps) {
      delayMicroseconds(50); // delay a small amount of time before refiring the steps to smooth things out
    } else {
      break;
    }
  } while(true);
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

// Update receive leds
// --------------------------------------
void UpdateReceiveLed(boolean value) {
  digitalWrite(ledPins[4], value);
}

// Step
// --------------------------------------
void Step(int stepPin, int dirPin, boolean dir) {

  digitalWrite(dirPin, dir);

  digitalWrite(stepPin, LOW);
  digitalWrite(stepPin, HIGH);
}

// Set all variables based on the data currently in the buffer
// --------------------------------------
void SetSliceVariables() {

  // set slice start pos to previous slice start plus previous delta
  leftStartPos = leftStartPos + long(leftDelta);
  rightStartPos = rightStartPos + long(rightDelta);

  if (moveDataLength < 2) {
    leftDelta = rightDelta = 0;
  } else {
    leftDelta = MoveDataGet();
    rightDelta = MoveDataGet();
  }
}

// Stop everything and blink the status led value times
// --------------------------------------
void Blink(char value) {
 int counts = value;
  if (counts<0) counts=-counts;

  UpdateReceiveLed(false);
  for(int i=0;i<counts;i++) {
   delay(1000);
   UpdateReceiveLed(true);
   delay(1000);
   UpdateReceiveLed(false);
    
  }
  delay(100000);
}

// Read serial data if its available
// --------------------------------------
void ReadSerialMoveData() {

  if(Serial.available()) {
    char value = Serial.read();
    
    // Check if this value is the sentinel reset value
    if (value == char(0x80)) {
      ResetMovementVariables();
      moveDataRequestPending = 0;
      moveDataLength = 0;
      UpdateReceiveLed(false);
      return;
    }

    MoveDataPut(value);
    moveDataRequestPending--;

    if (!moveDataRequestPending) {
      UpdateReceiveLed(false);
    }
  }
}

// Put a value onto the end of the move data buffer
// --------------------------------------
void MoveDataPut(char value) {
  

  int writePosition = moveDataStart + moveDataLength;
  if (writePosition >= MOVE_DATA_CAPACITY) {
    writePosition = writePosition - MOVE_DATA_CAPACITY;
  }

  moveData[writePosition] = value;

  if (moveDataLength == MOVE_DATA_CAPACITY) { // full, overwrite existing data
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
char MoveDataGet() {

  if (moveDataLength == 0) {
    return 0;
  }

  char result = moveData[moveDataStart];
  moveDataStart++;
  if (moveDataStart == MOVE_DATA_CAPACITY) {
    moveDataStart = 0;
  }
  moveDataLength--;

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
  UpdateReceiveLed(true);
}


