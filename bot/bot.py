import requests
import random
import time
import argparse
import os

def generate_nickname():
    first_names = ["Cool", "Super", "Mega", "Ultra", "Epic", "Lucky", "Silent", "Mighty"]
    last_names = ["Hero", "Warrior", "Ninja", "Samurai", "Wizard", "Knight", "Ranger", "Hunter"]
    nickname = f"{random.choice(first_names)}{random.choice(last_names)}{random.randint(0, 999)}"
    return nickname

def send_message(url):
    message = {
        "nickname": generate_nickname(),
        "text": "random"
    }
    response = requests.post(url, json=message)
    print(f"Message sent. Status: {response.status_code}")

def main():
    enable = os.environ.get("BOT_ENABLE", "false")

    if enable.lower() != "true":
        print("Bot is disabled")
        return

    rate = os.environ.get("BOT_MESSAGES_PER_MINUTE", 60)
    url = os.environ.get("BOT_SERVICE_URL", "http://localhost:3000/api/messages")

    interval = 60.0 / int(rate)

    while True:
        try:
            send_message(url)
        except Exception as e:
            print(f"Error: {e}")
        time.sleep(interval)

if __name__ == "__main__":
    main()
