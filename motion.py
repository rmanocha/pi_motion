from datetime import datetime, timedelta
import RPi.GPIO as GPIO
import time

GPIO.setwarnings(False)

MOTION_TIMEOUT = 5.0

MOTION_GPIO = 7
LIGHT_GPIO = 11

GPIO.setmode(GPIO.BCM) # see http://raspberrypi.stackexchange.com/a/12967

GPIO.setup(MOTION_GPIO, GPIO.IN) # this sets up the pin to act as an input
                                 # this is what we need to read the sensor

GPIO.setup(LIGHT_GPIO, GPIO.OUT) # this sets up the pin to act as an output

last_motion_ts = None

def turn_light_on():
    if GPIO.input(LIGHT_GPIO):
        GPIO.output(LIGHT_GPIO, GPIO.LOW) # this is what turns it on

def turn_light_off():
    if not GPIO.input(LIGHT_GPIO):
        GPIO.output(LIGHT_GPIO, GPIO.HIGH) # this is what turns if off

try:
    while True:
        if GPIO.input(MOTION_GPIO):
            last_motion_ts = datetime.now()
            #print "motion detected"
        else:
            #print "nothing"
            pass

        if last_motion_ts:
            if (datetime.now() - last_motion_ts).total_seconds() <= \
                                                                MOTION_TIMEOUT:
                #print "It's less than 5 seconds still"
                turn_light_on()
            else:
                #print "It's been more than 5 seconds now"
                turn_light_off()
                last_motion_ts = None
        #else:
        #    print "last_motion_ts is none"

        time.sleep(0.1)

except KeyboardInterrupt:
    print " quit"
    GPIO.cleanup()
