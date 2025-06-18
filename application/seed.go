package application

import (
	"fmt"

	"github.com/Leul-Michael/image-generation/model"
)

func (a *App) SeedCategories() error {
	categories := []model.Category{
		{Name: "Headshot", Description: "Professional portrait-style images, often used for profiles or resumes.", IsActive: true, Emoji: "🧑‍💼"},
		{Name: "Cartoonify", Description: "Transform real photos into fun cartoon-style illustrations.", IsActive: true, Emoji: "🎭"},
		{Name: "Lifestyle", Description: "Everyday scenes like a cozy family picnic or a sunny day at the park.", IsActive: true, Emoji: "🌞"},
		{Name: "Dream", Description: "Imaginative ideas like flying on a magical carpet or exploring a fantasy castle.", IsActive: true, Emoji: "💭"},
		{Name: "Fashion", Description: "Stylish outfits such as a colorful summer dress or a superhero costume.", IsActive: true, Emoji: "👗"},
		{Name: "Transport", Description: "Vehicles like a bright red fire truck or a cheerful hot air balloon.", IsActive: true, Emoji: "🚙"},
		{Name: "World Culture", Description: "Cultural themes like a Japanese cherry blossom festival or an African safari adventure.", IsActive: true, Emoji: "🌍"},
		{Name: "Stories", Description: "Storybook-inspired images like a pirate treasure hunt or a fairy tale forest.", IsActive: true, Emoji: "📖"},
		{Name: "Sport", Description: "Active scenes like a soccer game with friends or a fun bicycle race.", IsActive: true, Emoji: "⚽"},
		{Name: "Animals", Description: "Cute critters like a fluffy puppy or a playful dolphin.", IsActive: true, Emoji: "🐾"},
		{Name: "Colors", Description: "Vibrant designs like a rainbow-patterned kite or a sunset in warm hues.", IsActive: true, Emoji: "🌈"},
		{Name: "Ghibli Anime", Description: "Whimsical scenes inspired by Studio Ghibli, like a Totoro picnic or a Spirited Away train ride.", IsActive: true, Emoji: "🌀"},
		{Name: "Nature", Description: "Beautiful landscapes like a snowy mountain or a blooming flower garden.", IsActive: true, Emoji: "🌸"},
		{Name: "Food", Description: "Tasty treats like a giant ice cream sundae or a colorful fruit basket.", IsActive: true, Emoji: "🍰"},
		{Name: "Holidays", Description: "Festive moments like a Christmas tree lighting or a Halloween pumpkin patch.", IsActive: true, Emoji: "🎉"},
	}

	for _, category := range categories {
		var existingCategory model.Category
		result := a.DB.Where("name = ?", category.Name).First(&existingCategory)

		if result.Error != nil {
			if err := a.DB.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to create category %s: %w", category.Name, err)
			}
		}
	}

	return nil
}

func (a *App) SeedTrendingPrompts() error {
	var categories []model.Category
	if err := a.DB.Find(&categories).Error; err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	categoryMap := make(map[string]model.Category)
	for _, cat := range categories {
		categoryMap[cat.Name] = cat
	}

	trendingPrompts := []struct {
		prompt       string
		categoryName string
		useCount     int
	}{
		{"Professional headshot of a confident business man, studio lighting, crisp details, corporate attire", "Headshot", 0},
		{"Bioluminescent underwater scene with glowing jellyfish, deep ocean blues, magical lighting", "Nature", 0},
		{"Cherry blossom petals falling in slow motion, soft pink hues, dreamy spring atmosphere", "Nature", 0},
		{"Majestic lion with flowing mane, golden hour lighting, African savanna background, photorealistic", "Animals", 0},
		{"Enchanted library with floating books, magical glowing orbs, wizard's study atmosphere", "Dream", 0},
		{"Futuristic fashion model in iridescent outfit, neon city background, cyberpunk aesthetic", "Fashion", 0},
		{"Vintage train traveling through autumn mountains, steam locomotive, nostalgic journey", "Transport", 0},
	}

	for _, trendingPrompt := range trendingPrompts {
		category, exists := categoryMap[trendingPrompt.categoryName]
		if !exists {
			fmt.Printf("Warning: Category %s not found, skipping prompt\n", trendingPrompt.categoryName)
			continue
		}

		var existingPrompt model.TrendingPrompt
		result := a.DB.Where("prompt = ?", trendingPrompt.prompt).First(&existingPrompt)

		if result.Error != nil {
			newPrompt := model.TrendingPrompt{
				Prompt:     trendingPrompt.prompt,
				CategoryID: category.ID,
				UseCount:   trendingPrompt.useCount,
				IsActive:   true,
			}

			if err := a.DB.Create(&newPrompt).Error; err != nil {
				return fmt.Errorf("failed to create trending prompt: %w", err)
			}
		}
	}

	return nil
}
