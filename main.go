package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	vkapi "github.com/Dimonchik0036/vk-api"
)

// Constants

const token = ""
const categoriesFolder = "pictures/"

// Error handler

func handleError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// Init new user by token
func initializeUser(token string) (vkapi.LPChan, *vkapi.Client) {
	//client, err := vkapi.NewClientFromLogin("<username>", "<password>", vkapi.ScopeMessages)
	client, err := vkapi.NewClientFromToken(token)
	handleError(err)

	client.Log(true)

	if err := client.InitLongPoll(0, 2); err != nil {
		log.Panic(err)
	}

	updates, _, err := client.GetLPUpdatesChan(100, vkapi.LPConfig{25, vkapi.LPModeAttachments})
	handleError(err)

	return updates, client
}

// Get all categories from folder
func getImageCategories(folder string) []string {
	var categories []string

	file, err := os.Open(folder)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0)
	for _, name := range list {
		categories = append(categories, name)
	}

	return categories
}

// Get all images from categorie(folder)
func getAllImages(folder string) []string {
	var imageNames []string

	file, err := os.Open(folder)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0)
	for _, name := range list {
		imageNames = append(imageNames, name)
	}

	return imageNames
}

// Send images to user (categorie, quantity of images)
func sendImages(cat string, quantity int, client *vkapi.Client, update vkapi.LPUpdate) {
	var imageNames []string

	// Get all images
	imageNames = getAllImages(categoriesFolder + cat)

	// Shuffle images
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(imageNames), func(i, j int) { imageNames[i], imageNames[j] = imageNames[j], imageNames[i] })

	// If not enough images in categorie
	if len(imageNames) < quantity {
		quantity = len(imageNames)
	}

	// Send images to user
	for i := 0; i < quantity; i++ {

		// Path to current image
		var path string = categoriesFolder + cat + "/" + imageNames[i]

		// Read image from file that already exists
		existingImageFile, err := ioutil.ReadFile(path)
		handleError(err)

		client.SendPhoto(vkapi.NewDstFromUserID(update.Message.FromID), vkapi.FileBytes{Bytes: existingImageFile, Name: "picture.jpg"})
	}
}

// New message handler
func messageHandler(message string, client *vkapi.Client, update vkapi.LPUpdate) {

	// Get all categories from folder
	var categories []string = getImageCategories(categoriesFolder)

	// Clear string and set all letters to lowercase
	message = strings.ToLower(message)
	var words []string = strings.Fields(message)

	// Flags
	var isCat bool = false
	var isInfo bool = false

	// Input checking

	// If user has sent more than one word in message
	if len(words) > 2 {
		// Error sending
		client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), "Введите правильное количество картинок, например: \"люди 10\""))
		// If user has sent two words
	} else if len(words) == 2 {
		// Define categorie and quantity of images from user's message
		var userCat string = words[0]
		userQuantity, _ := strconv.Atoi(words[1])

		// If user has chosen more than 50 images from categorie
		if userQuantity > 50 {
			// Set categorie flag in true
			isCat = true
			// Error sending
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), "Картинок не должно быть больше 50"))
			// If quantity of images is correct
		} else {
			// Enumerate array of categories
			for _, cat := range categories {
				// Set categorie in lowercase
				cat = strings.ToLower(cat)
				cat = strings.TrimSpace(cat)

				// If categorie is defined
				if cat == userCat {
					// Set categorie flag in true
					isCat = true
					// Send images to user
					sendImages(cat, userQuantity, client, update)
				}
			}
		}
		// If user has sent less than 2 words
	} else if len(words) < 2 {
		// Define categorie of image
		var userCat string = words[0]

		// Check if user has wanted to see info
		if (message == "начать") || (message == "/info") {

			// Categories array with formatted categories
			var categoriesFormatted []string

			// Add , to each categorie
			for i := 0; i < len(categories); i++ {
				categoriesFormatted = append(categoriesFormatted, categories[i]+", ")
			}

			// Information for user
			var info string = fmt.Sprintf("Полный список доступных категорий: %v \n\nДля того чтобы получить рандомную картинку из категории, отправьте категорию картинки.\nЕсли вы хотите получить несколько картинок из категории, отправьте \"категория 10\", где 10 - количество картинок\n\nМаксимальное количество картинок в одном сообщении - 50", categoriesFormatted)
			isInfo = true

			// Send info
			client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), info))

			// Check for categorie in message
		} else {
			// Enumerate array of categories
			for _, cat := range categories {
				// Set categorie in lowercase
				cat = strings.ToLower(cat)
				cat = strings.TrimSpace(cat)

				// If categorie is defined
				if cat == userCat {
					isCat = true
					sendImages(cat, 1, client, update)
				}
			}
		}

	}
	// If categorie and info message were not defined
	if (isCat == false) && (isInfo == false) {
		// Send error to user
		client.SendMessage(vkapi.NewMessage(vkapi.NewDstFromUserID(update.Message.FromID), "Пожалуйста, введите категорию картинки, например: \"люди 10\""))
	}
}

func main() {
	// Define client and updates
	var updates vkapi.LPChan
	var client *vkapi.Client

	// Get clien and updates
	updates, client = initializeUser(token)

	// Check update for message
	for update := range updates {
		if update.Message == nil || !update.IsNewMessage() || update.Message.Outbox() {
			continue
		}

		log.Printf("%s", update.Message.String())
		// If message is defined
		if update.Message.Text != "" {
			// Run message handler
			messageHandler(update.Message.Text, client, update)

		}
	}
}
