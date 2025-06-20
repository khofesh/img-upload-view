import random

from PIL import Image

# random pixels
width, height = 4000, 4000
img = Image.new('RGB', (width, height))
pixels = []

for i in range(width * height):
    pixels.append((
        random.randint(0, 255),
        random.randint(0, 255), 
        random.randint(0, 255)
    ))

img.putdata(pixels)
img.save('large_random.jpg', 'JPEG', quality=100)
print("Created large_random.jpg")
