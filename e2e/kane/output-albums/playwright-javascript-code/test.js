import testmu, { resolveVar, setVar } from '@testmuai/playwright-bindings';
import { expect } from '@playwright/test';

testmu.configure({
  "build": "b4e143b6-e238-4df3-99ff-70afc653e5e9",
  "name": "Set Up Photoview and Verify Vacation Album",
  "tcId": "TC-197",
  "network": (process.env.NETWORK ?? "false").toLowerCase() === "true",
  "autoHealVersion": "AH2",
  "defaultActionTimeoutMs": 10000,
  "defaultNavigationTimeoutMs": 60000,
  "kaneRunV4": true
});

testmu.test("Set Up Photoview and Verify Vacation Album", async (page) => {
  let element_0, element_1, element_2, element_3, element_4, element_5, element_6, element_7;
  await testmu.step("Navigate to http://localhost:8000/", { instructionId: "5a727963-65f6-43cf-8823-a30236f16ed0" }, async () => {
    await page.goto("http://localhost:8000/");
  });
  
  await testmu.step("Waiting for the queued plan to run", { instructionId: "3bc014f6-a9cc-401c-ac04-24227107c625" }, async () => {
    await page.waitForTimeout(1000);
  });
  
  await testmu.step("Type \"admin\" into \"USERNAME\"", { instructionId: "b728903c-8467-43fc-8272-c032444ab346" }, async () => {
    element_0 = page.locator("internal:role=textbox[name=\"Username\"i]");
    await element_0.click();
    await element_0.fill("admin");
  });
  
  await testmu.step("Type \"admin12345\" into \"PASSWORD Show password\"", { instructionId: "c961afc5-c3fe-44d4-86f9-87e315ef11bb" }, async () => {
    element_1 = page.locator("internal:role=textbox[name=\"Password Show password\"i]");
    await element_1.click();
    await element_1.fill("admin12345");
  });
  
  await testmu.step("Waiting while queuing the plan step to type PHOTO PATH", { instructionId: "cbca7988-d5b7-4627-8919-b0e90b217d96" }, async () => {
    await page.waitForTimeout(1000);
  });
  
  await testmu.step("Type \"/photos\" into \"PHOTO PATH\"", { instructionId: "f956317d-8d1d-4811-9d63-d84317c78818" }, async () => {
    element_2 = page.locator("internal:role=textbox[name=\"Photo path\"i]");
    await element_2.click();
    await element_2.fill("/photos");
  });
  
  await testmu.step("Waiting for the completed setup form staying populated", { instructionId: "ef4edc3a-d9c0-4066-878f-b8b020235767" }, async () => {
    await page.waitForTimeout(500);
  });
  
  await testmu.step("Clicking Setup Photoview button", { instructionId: "444797cd-4cbb-4f04-9cd8-86d13541b1f5" }, async () => {
    element_3 = page.locator("internal:role=button[name=\"Setup Photoview\"i]");
    await element_3.click();
  });
  
  await testmu.step("Waiting for the Timeline page to finish loading", { instructionId: "a461f18b-3012-464c-8fad-e993de6094d5" }, async () => {
    await page.waitForTimeout(1000);
  });
  
  await testmu.step("Reading the current browser URL", { instructionId: "7ec6a78f-9b2e-470a-9eb5-d66167e79644" }, async () => {
    setVar("url_contains_timeline", page.url());
  });
  
  await testmu.step("Assertion check", { instructionId: "2b2f9e9b-27f7-4822-9fe8-fb20a1f067f0" }, async () => {
    await testmu.verifyAssertion(page, "Assertion check", {"operator": ["equals"], "assertion_operands": [], "left_operand": null, "right_operand": null, "operands": [], "sub_results": [{"description": "The current browser URL contains \"/timeline\".", "passed": true, "operator": "contains", "transforms": [], "json_path": null, "expected": "/timeline", "extracted_value": "{{url_contains_timeline}}", "store_key": "url_contains_timeline", "variable_refs": {"{{url_contains_timeline}}": "http://localhost:8000/timeline"}}], "sub_checks": [{"description": "The current browser URL contains \"/timeline\".", "store_key": "url_contains_timeline", "expected_value": "/timeline", "extracted_value": "{{url_contains_timeline}}", "operator": "contains", "transforms": []}], "composite_operator": "and", "claim": "the URL contains \"/timeline\""});
  });
  
  await testmu.step("Clicking the Settings link in the left sidebar", { instructionId: "16367d60-086d-4c36-99c9-a611292e2da0" }, async () => {
    element_4 = page.locator("internal:role=link[name=\"Settings\"i]");
    await element_4.click();
  });
  
  await testmu.step("Scanning all users by clicking Scan all users button", { instructionId: "279ab370-9079-475c-91ff-a949575f6221" }, async () => {
    element_5 = page.locator("internal:role=button[name=\"Scan all users\"i]");
    await element_5.click();
  });
  
  await testmu.step("Waiting for the user scan to finish", { instructionId: "95bd98da-cb91-443c-a8db-5b93e7e731ba" }, async () => {
    await page.waitForTimeout(10000);
  });
  
  await testmu.step("Checking whether the Settings page is shown", async () => {
    await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
    setVar("settings_page_still_shown", await testmu.textualAnalyzer(page, { wrappedJs: "(els) => {\n  const __m = {22: 0};\n  const el = (i) => els[__m[i]];\n  const __v = (el(22).textContent.trim() === 'User preferences');\n  return (typeof __v === 'boolean' ? String(__v) : __v);\n}", locators: ["internal:role=heading[name=\"User preferences\"i]"], query: "whether the Settings page user preferences heading is present", expectedValue: "true", needsUnitConversion: false, operator: "equals", transforms: [], condition: "the Settings page is still shown" }));
  });
  
  await testmu.step("Assertion check", { instructionId: "50d36a3d-79af-4fe8-9ba7-443c52c7b9fa" }, async () => {
    await testmu.verifyAssertion(page, "Assertion check", {"operator": ["equals"], "assertion_operands": [], "left_operand": null, "right_operand": null, "operands": [], "sub_results": [{"description": "the Settings page is still shown", "passed": true, "operator": "equals", "transforms": [], "json_path": null, "expected": "true", "extracted_value": "{{settings_page_still_shown}}", "store_key": "settings_page_still_shown", "variable_refs": {"{{settings_page_still_shown}}": "true"}}], "sub_checks": [{"description": "the Settings page is still shown", "store_key": "settings_page_still_shown", "expected_value": "true", "extracted_value": "{{settings_page_still_shown}}", "operator": "equals", "transforms": []}], "composite_operator": "and", "claim": "the Settings page is still shown"});
  });
  
  await testmu.step("Clicking Albums navigation link in the left sidebar", { instructionId: "6f73ecdb-0050-43d0-a37d-fff32417c457" }, async () => {
    element_6 = page.locator("internal:role=link[name=\"Albums\"i]");
    await element_6.click();
  });
  
  await testmu.step("Checking whether the Vacation album is listed", async () => {
    await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
    setVar("vacation_album_not_listed", await testmu.textualAnalyzer(page, { wrappedJs: "(els) => {\n  const __m = {35: 0};\n  const el = (i) => els[__m[i]];\n  const __v = (el(35) !== null);\n  return (typeof __v === 'boolean' ? String(__v) : __v);\n}", locators: ["internal:role=link[name=\"Vacation\"i]"], query: "whether the Vacation album is listed in the albums page", expectedValue: "false", needsUnitConversion: false, operator: "equals", transforms: [], condition: "The \"Vacation\" album is not listed yet." }));
  });
  
  if (await testmu.evaluateBranch([{"description": "The \"Vacation\" album is not listed yet.", "store_key": "vacation_album_not_listed", "expected_value": "false", "stored_value": "{{vacation_album_not_listed}}", "operator": "equals", "transforms": []}], "and")) {
    page = await testmu.executeKaneCli("open the \"photos\" album", page);
  }
  
  await testmu.step("Clicking Vacation album link in the albums grid", { instructionId: "909cd5a0-8338-44cb-95ea-2fd086e8f53b" }, async () => {
    element_7 = page.locator("internal:role=link[name=\"Vacation\"i]");
    await element_7.click();
  });
  
  await testmu.step("Waiting for the Vacation album page to finish loading", { instructionId: "c7f4c525-e1fe-40cb-b1ef-7ead5049bbe6" }, async () => {
    await page.waitForTimeout(1000);
  });
  
  await testmu.step("Reading the album page title", async () => {
    await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
    setVar("album_page_title", await testmu.textualAnalyzer(page, { wrappedJs: "(els) => {\n  const __m = {39: 0};\n  const el = (i) => els[__m[i]];\n  const __v = (el(39).textContent.trim());\n  return (typeof __v === 'boolean' ? String(__v) : __v);\n}", locators: ["internal:testid=[data-testid=\"Layout\"s] >> internal:role=heading[name=\"Vacation\"i]"], query: "the album page title", expectedValue: "Vacation", needsUnitConversion: false, operator: "equals", transforms: ["strip"], condition: "The album page shows \"Vacation\"." }));
  });
  
  await testmu.step("Reading the album photo count", async () => {
    await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
    setVar("album_photo_count", await testmu.textualAnalyzer(page, { wrappedJs: "(els) => {\n  const __m = {42: 0};\n  const el = (i) => els[__m[i]];\n  const __v = (el(42).querySelectorAll('img[src*=\"/api/photo/\"]').length);\n  return (typeof __v === 'boolean' ? String(__v) : __v);\n}", locators: ["internal:testid=[data-testid=\"photo-gallery-wrapper\"s]"], query: "the number of photos displayed in the album gallery", expectedValue: "3", needsUnitConversion: false, operator: "equals", transforms: ["string_to_float"], condition: "The album page shows 3 photos." }));
  });
  
  await testmu.step("Assertion check", { instructionId: "6576db4a-7c68-45e3-b3df-1b158bd4ea9d" }, async () => {
    await testmu.verifyAssertion(page, "Assertion check", {"operator": ["equals"], "assertion_operands": [], "left_operand": null, "right_operand": null, "operands": [], "sub_results": [{"description": "The album page shows \"Vacation\".", "passed": true, "operator": "equals", "transforms": ["strip"], "json_path": null, "expected": "Vacation", "extracted_value": "{{album_page_title}}", "store_key": "album_page_title", "variable_refs": {"{{album_page_title}}": "Vacation"}}, {"description": "The album page shows 3 photos.", "passed": true, "operator": "equals", "transforms": ["string_to_float"], "json_path": null, "expected": "3", "extracted_value": "{{album_photo_count}}", "store_key": "album_photo_count", "variable_refs": {"{{album_photo_count}}": "3"}}], "sub_checks": [{"description": "The album page shows \"Vacation\".", "store_key": "album_page_title", "expected_value": "Vacation", "extracted_value": "{{album_page_title}}", "operator": "equals", "transforms": ["strip"]}, {"description": "The album page shows 3 photos.", "store_key": "album_photo_count", "expected_value": "3", "extracted_value": "{{album_photo_count}}", "operator": "equals", "transforms": ["string_to_float"]}], "composite_operator": "and", "claim": "Verify the album page shows \"Vacation\" and 3 photos."});
  });
});

testmu.run();