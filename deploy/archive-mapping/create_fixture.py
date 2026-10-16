"""Generate harmless two-page originals for isolated mapping acceptance."""
from pathlib import Path
import sys
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4

root = Path(sys.argv[1]).resolve()
for revision, region in [(1, "国家"), (1, "地方"), (2, "国家")]:
    destination = root / f"v{revision}" / "原件" / region / "映射验收.pdf"
    destination.parent.mkdir(parents=True, exist_ok=True)
    document = canvas.Canvas(str(destination), pagesize=A4, invariant=True)
    for page in (1, 2):
        document.setFont("Helvetica-Bold", 20)
        document.drawString(50, 780, "Archive Mapping Acceptance")
        document.setFont("Helvetica", 12)
        document.drawString(50, 750, f"Revision {revision} / Region {'National' if region == '国家' else 'Local'} / Page {page} of 2")
        document.drawString(50, 710, "Original PDF, stored without OCR or model processing.")
        for row in range(4):
            y = 650 - row * 35
            document.rect(50, y, 490, 35)
            document.line(260, y, 260, y + 35)
            document.drawString(60, y + 12, f"Evidence row {row + 1}")
            document.drawString(270, y + 12, f"V{revision}-P{page}-R{row + 1}")
        document.showPage()
    document.save()
    print(destination)
