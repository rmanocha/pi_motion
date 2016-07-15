import RPi.GPIO as GPIO
import time

MOTION_GPIO=7
GPIO.setmode(GPIO.BCM)
GPIO.setup(MOTION_GPIO, GPIO.IN)

try:
    while True:
        if GPIO.input(MOTION_GPIO):
            print "motion detected"
        else:
            print "nothing"
                time.sleep(0.1)

except KeyboardInterrupt:
    print " quit"
        GPIO.cleanup()
