#!/usr/bin/env python3
"""Create synthetic PDF fixtures. No historical documents are read."""
from pathlib import Path
from reportlab.pdfgen import canvas
from reportlab.lib.utils import ImageReader
from pypdf import PdfReader,PdfWriter
from PIL import Image,ImageDraw
import io
out=Path(__file__).resolve().parents[2]/'runtime/pdf-fixtures';out.mkdir(parents=True,exist_ok=True)
def make(name,count,blank=False,scan=False):
 c=canvas.Canvas(str(out/name),pagesize=(400,500))
 for i in range(count):
  if not(blank and i==1):
   lines=[f'FULLPAGE TEST {i+1} OF {count}',f'UNIQUE END MARKER ZEBRA{count:02d}PAGE{i+1:02d}', 'Record | Description | Value',f'{i+1} | Independent validation fixture | 17']
   if scan:
    image=Image.new('RGB',(800,1000),'white');d=ImageDraw.Draw(image)
    for j,text in enumerate(lines):d.text((50,100+j*90),text,fill='black',font_size=28)
    c.drawImage(ImageReader(image),0,0,width=400,height=500)
   else:
    c.setFont('Helvetica',14)
    for j,text in enumerate(lines):c.drawString(25,430-j*55,text)
  c.showPage()
 c.save()
make('single.pdf',1);make('sixteen.pdf',16);make('long24.pdf',24);make('blank-page.pdf',3,blank=True);make('scan.pdf',1,scan=True)
writer=PdfWriter();writer.append(str(out/'single.pdf'));writer.encrypt('independent-fixture-password');writer.write(str(out/'encrypted.pdf'))
(out/'corrupt.pdf').write_bytes(b'%PDF-1.4\nthis is deliberately damaged\n%%EOF')
for p in out.glob('*.pdf'):print(p.name,p.stat().st_size)
