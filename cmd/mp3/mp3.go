package main

import (
	"os"
	"path/filepath"

	"github.com/dawidjelenkowski/aidevs3go/internal/logging"
	"github.com/dawidjelenkowski/aidevs3go/internal/transcribe"
	"github.com/dawidjelenkowski/aidevs3go/internal/utils"
	"github.com/dawidjelenkowski/aidevs3go/internal/vertex"
	"github.com/rs/zerolog/log"
)

// constants
const location = "us-central1"
const project = "avid-truth-426717-v0"
const model = "gemini-2.0-flash-exp"

// Define the input and output directories
const inputDir = "documents/przesluchania"
const outputDir = "downloads/audio"

func main() {
	logging.Setup("mp3")
	log.Info().Msg("Starting mp3 processing")

	APIKEY, err := utils.GetAPIKey("gemini-api-key") //gemini-api-key or openai-api-key
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get Gemini API key")
	}

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal().Err(err).Str("path", outputDir).Msg("Failed to create output directory")
		return
	}

	// Transcribe audio files
	err = transcribe.TranscribeAudioFiles(APIKEY, inputDir, outputDir, "gemini") // "gemini" or "whisper"
	if err != nil {
		log.Error().Err(err).Msg("Error during audio transcription")
	}
	log.Info().Msg("Audio transcription completed. Check the logs for details.")

	// create prompt
	var prompt string
	files, err := os.ReadDir(outputDir)
	if err != nil {
		log.Error().Err(err).Str("directory", outputDir).Msg("Failed to read transcript directory")
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			filePath := filepath.Join(outputDir, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Error().Err(err).Str("filePath", filePath).Msg("Failed to read transcription file")
				continue
			}
			prompt += string(content) + "\n\n"
		}
	}

	// Ask Gemini the question
	system := "Odpowiedz zwięźle na pytanie: na jakiej ulicy znajduje się uczelnia, na której wykłada Andrzej Maj?"

	// ask gemini
	// geminiAnswer, err := gemini.AskGemini(&gemini.GeminiConfig{
	// 	GeminiAPIKey: APIKEY,
	// 	Model:        model,
	// 	System:       system,
	// 	Prompt:       prompt,
	// })

	vertexAnswer, err := vertex.AskVertex(&vertex.VertexConfig{
		Project:  project,
		Location: location,
		Model:    model,
		System:   system,
		Prompt:   prompt,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to get answer from Gemini")
	} else {
		log.Info().Msg("Gemini answered the question")

		// Send the answer
		taskName := "mp3"
		err = utils.SendAnswer(vertexAnswer, taskName)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send answer")
		} else {
			log.Info().Msg("Answer sent successfully")
		}
	}
}
