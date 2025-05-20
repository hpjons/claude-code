import strutils, random

randomize()
let answer = rand(3)
while true:
    echo "I have a number from 0 to 10, what is it? "
    let guess = parseInt(stdin.readline)

    if guess < answer:
        echo "Too low, try again"
    elif guess > answer:
        echo "Too high, try again"
    else:
        echo "Correct!"
        break

block busyLoops:
    while true:
        while true:
            break busyLoops    