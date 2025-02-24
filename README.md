<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/github_username/repo_name">
    <img src="media/logo.png" alt="Logo" width="80" height="80">
  </a>

<h3 align="center">Luodingo Telegram Bot</h3>

  <p align="center">
    Luodingo Telegram Bot helps you learn new words with interactive flashcards. Test your memory, track progress, and expand your vocabulary effortlessly!
    <br />
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#try-luodingo-for-yourself">Try Luodingo for Yourself</a></li>
        <li><a href="#run-locally-with-docker-compose">Run Locally with Docker Compose</a></li>
      </ul>
    </li>
    <li><a href="#features">Features</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

Luodingo is a Telegram bot designed to help you learn new words through flashcards. You can create custom flashcards, organize them into decks, and review them at your own pace. Luodingo helps reinforce your learning with spaced repetition techniques, making vocabulary acquisition more efficient and fun! 




<!-- GETTING STARTED -->
## Getting Started  

### Try Luodingo for Yourself  
Luodingo is available on Telegram! Start learning now by chatting with [@luodingobot](https://t.me/luodingobot).  
For a full list of features, check out the [Features](#features) section.  

### Run Locally with Docker Compose  
You can easily run Luodingo locally using Docker Compose by following these steps:  

#### 1. Download the `docker-compose.yaml` file  
Open your terminal and run:  

```sh
wget https://raw.githubusercontent.com/dafraer/luodingo-tg-bot/refs/heads/main/docker-compose.yaml
```  

#### 2. Set Up Architechture and  Environment Variables  
- Replace `<your_telegram_bot_token>` with your bot token (Get it from [@BotFather](https://t.me/BotFather)).  
- Choose the correct image tag based on your system architecture:  
  - **For x86_64 (AMD64):** Use `3.0-amd64`  
  - **For ARM64 (e.g., Raspberry Pi):** Use `3.0-arm64`  

#### 3. Start the Bot  
Run the following command to start Luodingo in the background:  

```sh
sudo docker-compose up -d
```  

Now your bot should be up and running locally! 🚀


<!-- FEATURES -->
## Features  

Luodingo bot offers a variety of features to help you learn efficiently:  

- 📚 **Create a New Deck** – Use the `/new_deck` command to create a new flashcard deck.  
- ➕ **Add Cards** – Add new flashcards to a deck using the `/add_cards` command.  
- 📋 **List All Decks** – View all your created decks with the `/my_decks` command.  
- 🔍 **View Deck Cards** – See all flashcards in a specific deck using `/my_cards`.  
- ❌ **Delete a Deck** – Remove an entire deck with the `/delete_deck` command.  
- 🗑 **Delete a Card** – Remove a specific card using the `/delete_card` command.  
- 🎓 **Study Mode** – Use the `/study_deck` command to review and test yourself.  
- 🔄 **Add Reverse Cards** – Automatically generate reverse flashcards for better learning.  

Luodingo makes vocabulary building simple and interactive! 🚀  



<!-- CONTACT -->
## Contact

Kamil Nuriev- [telegram](https://t.me/dafraer) - kdnuriev@gmail.com




