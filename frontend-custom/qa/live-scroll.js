async(page)=>{
 const input=page.getByRole('spinbutton',{name:'PDF 页码'})
 const before=Number(await input.inputValue())
 await page.locator('.archival-pdf-scroll').hover()
 await page.mouse.wheel(0,1500)
 await page.waitForFunction(n=>Number(document.querySelector('[aria-label="PDF 页码"]').value)>n,before)
 console.log({before,after:Number(await input.inputValue()),renderedCanvases:await page.locator('.archival-pdf-scroll canvas').count()})
 await page.screenshot({path:'output/playwright/archival-live-pdf-scrolled.png'})
}
