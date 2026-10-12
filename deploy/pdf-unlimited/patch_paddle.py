#!/usr/bin/env python3
"""Apply a version-checked Paddle serving patch; invoke after ubuntu login + su."""
import pathlib, shutil, py_compile
root=pathlib.Path('/data/ocr-services/paddleocr-3.6-vl1.6')
project=pathlib.Path('/home/ubuntu/RAG_Weknora')
p=root/'.venv/lib/python3.12/site-packages/paddlex/inference/serving/basic_serving/_pipeline_apps/paddleocr_vl.py'
s=p.read_text()
if 'weknora_full_pages_v1' not in s:
 s=s.replace('from fastapi import FastAPI','from fastapi import FastAPI, HTTPException\n    from fastapi.responses import JSONResponse')
 s=s.replace('    ocr_common.update_app_context(ctx)','''    ocr_common.update_app_context(ctx)
    # weknora_full_pages_v1: observable effective config and live requests.
    active_requests = [0]
    @app.middleware("http")
    async def full_page_request_count(request, call_next):
        active = request.url.path == INFER_ENDPOINT
        if active: active_requests[0] += 1
        try:
            return await call_next(request)
        finally:
            if active: active_requests[0] -= 1

    @app.get("/full-page-status")
    async def full_page_status():
        return {"max_num_input_imgs": ctx.extra["max_num_input_imgs"],
                "active_requests": active_requests[0], "patch": "weknora_full_pages_v1"}
''')
 marker='        layout_parsing_results: List[Dict[str, Any]] = []'
 assert s.count(marker)==1
 s=s.replace(marker,'''        result = list(result)
        if len(result) != len(images):
            raise HTTPException(status_code=500, detail="OCR result count mismatch")
'''+marker)
 s=s.replace('            pruned_res = common.prune_result(item.json["res"])','''            pruned_res = common.prune_result(item.json["res"])
            pruned_res["page_index"] = i''',1)
 marker='''        return AIStudioResultResponse[InferResult](
            logId=log_id,
            result=InferResult(
                layoutParsingResults=layout_parsing_results,
                dataInfo=data_info,
            ),
        )'''
 assert marker in s
 s=s.replace(marker,'''        response = AIStudioResultResponse[InferResult](
            logId=log_id,
            result=InferResult(layoutParsingResults=layout_parsing_results, dataInfo=data_info),
        ).model_dump()
        response["result"]["dataInfo"]["processedPages"] = len(layout_parsing_results)
        return JSONResponse(content=response)''')
 candidate=project/'runtime/paddleocr_vl.patched.py';candidate.write_text(s)
 py_compile.compile(str(candidate),doraise=True)
 shutil.copy2(candidate,p)
config=root/'config/PaddleOCR-VL-1.6.yaml'
s=config.read_text()
import re
s,n=re.subn(r'(?m)^(\s*max_num_input_imgs:)\s*[^\n]+$',r'\1 null',s)
assert n==1,'explicit max_num_input_imgs key required'
config.write_text(s)
(project/'config/PaddleOCR-VL-1.6-unlimited.yaml').write_text(s)
print('Paddle patch installed; max_num_input_imgs: null')
