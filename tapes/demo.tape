# Hoptimist CLI Demo Tape

Output assets/demo.webm

# Set up shell and speed
Set Shell "bash"
Set TypingSpeed 100ms
Set PlaybackSpeed 1
Set Height 800
Set Theme "Catppuccin Mocha"

# Start the application
Hide
Type "./hopcli" Enter
Sleep 2s
Show

# Show latest products
# (Assuming "Latest" is the first option and selected by default, or press Enter to select)
Enter
Sleep 3s

# Navigate pagination (next page)
Type "n"
Sleep 1s
Type "n"
Sleep 3s

# Navigate pagination (previous page)
Type "p"
Sleep 3s

# Go back to main menu
Type "h"
Sleep 2s

# Navigate to Categories (move down then enter)
Down Enter
Sleep 3s

# Select a category (e.g., move down once, then enter)
# This might need adjustment based on actual categories
Down Enter
Sleep 3s

# Navigate pagination in category products
Type "n"
Sleep 3s

# Select a product (e.g., move down once, then enter)
# This might need adjustment
Down Enter
Sleep 4s

# Go back from product view
Type "h"
Sleep 2s

# Go back from category products view
Type "h"
Sleep 2s

# Go back from categories view
Type "h"
Sleep 2s

# Quit
Type "q"
Enter # VHS needs an Enter after q if q itself doesn't exit immediately in the app
Sleep 1s
