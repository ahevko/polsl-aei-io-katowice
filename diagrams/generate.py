from pathlib import Path
import subprocess
from PIL import Image, ImageDraw

output_dir = Path("out")
output_dir.mkdir(parents=True, exist_ok=True)

# ludzik UML (aktor) jako obrazek
im = Image.new("RGBA", (60, 110), (255, 255, 255, 0))  # rysowany x2 niżej
d = ImageDraw.Draw(im)
c = (168, 0, 54, 255)
d.ellipse((18, 2, 42, 26), outline=c, width=3, fill=(254, 254, 206, 255))
d.line((30, 26, 30, 68), fill=c, width=3)
d.line((8, 40, 52, 40), fill=c, width=3)
d.line((30, 68, 10, 104), fill=c, width=3)
d.line((30, 68, 50, 104), fill=c, width=3)
im = im.resize((120, 220), Image.LANCZOS)
im.save(output_dir / "actor.png")

for dot_file in Path(".").glob("*.dot"):
    output_file = output_dir / f"{dot_file.stem}.png"
    subprocess.run(
        ["dot", "-Tpng", str(dot_file), "-o", str(output_file), "-Gdpi=220"],
        check=True,
    )

# Process all PlantUML PUML files in the current directory
for puml_file in Path(".").glob("*.puml"):
    subprocess.run(
        ["java", "-jar", "plantuml.jar", "-tpng", "-o", str(output_dir.resolve()), str(puml_file)],
        check=True,
    )

print("ok")
