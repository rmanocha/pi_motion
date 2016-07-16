from datetime import datetime, timedelta
import RPi.GPIO as GPIO
import time

MOTION_TIMEOUT = 5.0

MOTION_GPIO=7
GPIO.setmode(GPIO.BCM)
GPIO.setup(MOTION_GPIO, GPIO.IN)

last_motion_ts = None

try:
    while True:
        if GPIO.input(MOTION_GPIO):
            last_motion_ts = datetime.now()
            print "motion detected"
        else:
            print "nothing"

        if last_motion_ts:
            if (datetime.now() - last_motion_ts).total_seconds() <= \
                                                                MOTION_TIMEOUT:
                print "It's less than 5 seconds still"
            else:
                print "It's been more than 5 seconds now"
                last_motion_ts = None
        else:
            print "last_motion_ts is none"

        time.sleep(0.1)

except KeyboardInterrupt:
    print " quit"
    GPIO.cleanup()
