import testmu, { resolveVar, setVar } from '@testmuai/playwright-bindings';
import { expect } from '@playwright/test';

testmu.configure({
  "build": "d5a5a025-ae3b-4e48-a3cd-919eeb5e0836",
  "name": "Complete Initial Setup for Photoview",
  "tcId": "TC-196",
  "network": (process.env.NETWORK ?? "false").toLowerCase() === "true",
  "autoHealVersion": "AH2",
  "defaultActionTimeoutMs": 10000,
  "defaultNavigationTimeoutMs": 60000,
  "kaneRunV4": true
});

testmu.test("Complete Initial Setup for Photoview", async (page) => {
  let element_0, element_1, element_2, element_3;
  await testmu.step("Navigate to http://localhost:8000/", { instructionId: "70ca99d7-9878-47de-8452-6672bd52935e" }, async () => {
    await page.goto("http://localhost:8000/");
  });
  
  await testmu.step("Reading the current page URL", { instructionId: "32c597b2-a905-4016-ab99-c09c5afd745b" }, async () => {
    setVar("initial_setup_url_check", page.url());
  });
  
  await testmu.step("Reading the initial setup page heading", async () => {
    await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
    setVar("initial_setup_heading_check", await testmu.textualAnalyzer(page, { wrappedJs: "(els) => {\n  const __m = {2: 0};\n  const el = (i) => els[__m[i]];\n  const __v = (el(2).textContent.trim());\n  return (typeof __v === 'boolean' ? String(__v) : __v);\n}", locators: ["internal:role=heading[name=\"Initial Setup\"i]"], query: "the main setup page heading text", expectedValue: "Initial Setup", needsUnitConversion: false, operator: "equals", transforms: ["strip"], condition: "the page shows \"Initial Setup\"" }));
  });
  
  await testmu.step("Reading the setup form field labels", async () => {
    await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
    setVar("initial_setup_fields_check", await testmu.textualAnalyzer(page, { wrappedJs: "(els) => {\n  const __m = {4: 0, 6: 1, 11: 2};\n  const el = (i) => els[__m[i]];\n  const __v = ([el(4).textContent.trim(),el(6).textContent.trim(),el(11).textContent.trim()].join('|'));\n  return (typeof __v === 'boolean' ? String(__v) : __v);\n}", locators: ["label >> internal:has-text=\"Username\"i", "label >> internal:has-text=\"Password\"i", "label >> internal:has-text=\"Photo path\"i"], query: "the Username, Password, and Photo path field labels in the setup form", expectedValue: "username|password|photo path", needsUnitConversion: false, operator: "equals", transforms: ["lowercase"], condition: "the page shows \"Username\", \"Password\" and \"Photo path\" fields" }));
  });
  
  await testmu.step("Assertion check", { instructionId: "be75bae5-4827-4c70-9e34-37e7539b4a50" }, async () => {
    await testmu.verifyAssertion(page, "Assertion check", {"operator": ["equals"], "assertion_operands": [], "left_operand": null, "right_operand": null, "operands": [], "sub_results": [{"description": "the URL contains \"/initialSetup\"", "passed": true, "operator": "contains", "transforms": [], "json_path": null, "expected": "/initialSetup", "extracted_value": "{{initial_setup_url_check}}", "store_key": "initial_setup_url_check", "variable_refs": {"{{initial_setup_url_check}}": "http://localhost:8000/initialSetup"}}, {"description": "the page shows \"Initial Setup\"", "passed": true, "operator": "equals", "transforms": ["strip"], "json_path": null, "expected": "Initial Setup", "extracted_value": "{{initial_setup_heading_check}}", "store_key": "initial_setup_heading_check", "variable_refs": {"{{initial_setup_heading_check}}": "Initial Setup"}}, {"description": "the page shows \"Username\", \"Password\" and \"Photo path\" fields", "passed": true, "operator": "equals", "transforms": ["lowercase"], "json_path": null, "expected": "username|password|photo path", "extracted_value": "{{initial_setup_fields_check}}", "store_key": "initial_setup_fields_check", "variable_refs": {"{{initial_setup_fields_check}}": "Username|Password|Photo path"}}], "sub_checks": [{"description": "the URL contains \"/initialSetup\"", "store_key": "initial_setup_url_check", "expected_value": "/initialSetup", "extracted_value": "{{initial_setup_url_check}}", "operator": "contains", "transforms": []}, {"description": "the page shows \"Initial Setup\"", "store_key": "initial_setup_heading_check", "expected_value": "Initial Setup", "extracted_value": "{{initial_setup_heading_check}}", "operator": "equals", "transforms": ["strip"]}, {"description": "the page shows \"Username\", \"Password\" and \"Photo path\" fields", "store_key": "initial_setup_fields_check", "expected_value": "username|password|photo path", "extracted_value": "{{initial_setup_fields_check}}", "operator": "equals", "transforms": ["lowercase"]}], "composite_operator": "and", "claim": "the URL contains \"/initialSetup\" and the page shows \"Initial Setup\" with \"Username\", \"Password\" and \"Photo path\" fields"});
  });
  
  await testmu.step("Waiting for the next step in the queued plan", { instructionId: "93bf7be6-058a-427d-b3a1-aac403e77a1d" }, async () => {
    await page.waitForTimeout(1000);
  });
  
  await testmu.step("Type \"admin\" into \"USERNAME\"", { instructionId: "2e305cf3-0770-4901-9d38-bb6c4597c890" }, async () => {
    element_0 = page.locator("internal:role=textbox[name=\"Username\"i]");
    await element_0.click();
    await element_0.fill("admin");
  });
  
  await testmu.step("Type \"admin12345\" into \"PASSWORD Show password\"", { instructionId: "7c447dfb-ce7b-4c00-b466-37804d9928f7" }, async () => {
    element_1 = page.locator("internal:role=textbox[name=\"Password Show password\"i]");
    await element_1.click();
    await element_1.fill("admin12345");
  });
  
  await testmu.step("Waiting for the queued plan to start", { instructionId: "1fa26aec-d405-4b25-a478-47810df8c164" }, async () => {
    await page.waitForTimeout(1000);
  });
  
  await testmu.step("Type \"/photos\" into \"PHOTO PATH\"", { instructionId: "32aa5ecc-45eb-48d8-b450-20a3d64e1ca1" }, async () => {
    element_2 = page.locator("internal:role=textbox[name=\"Photo path\"i]");
    await element_2.click();
    await element_2.fill("/photos");
  });
  
  await testmu.step("Clicking Setup Photoview submit button", { instructionId: "f2271204-ecc9-46fd-96a6-c4e9f07639fe" }, async () => {
    element_3 = page.locator("internal:role=button[name=\"Setup Photoview\"i]");
    await element_3.click();
  });
  
  await testmu.step("Waiting 1000 ms for the Timeline page to finish loading", { instructionId: "1df20b26-292c-4b2f-b66c-c9b5cf35d85f" }, async () => {
    await page.waitForTimeout(1000);
  });
  
  await testmu.step("Reading the current page URL", { instructionId: "0543099f-a6a4-419b-958d-7eea03cf0361" }, async () => {
    setVar("timeline_url_check", page.url());
  });
  
  await testmu.step("Checking the sidebar navigation links", async () => {
    await page.waitForLoadState("domcontentloaded", { timeout: 5000 });
    setVar("sidebar_navigation_check", await testmu.textualAnalyzer(page, { wrappedJs: "(els) => {\n  const __m = {7: 0};\n  const el = (i) => els[__m[i]];\n  const __v = (['Timeline','Albums','People','Settings'].every(label => Array.from(el(7).querySelectorAll('a')).some(link => link.textContent.trim() === label)));\n  return (typeof __v === 'boolean' ? String(__v) : __v);\n}", locators: ["internal:text=\"TimelineAlbumsPeopleSettings\"i"], query: "whether the sidebar navigation includes Timeline, Albums, People, and Settings", expectedValue: "true", needsUnitConversion: false, operator: "equals", transforms: [], condition: "the sidebar shows \"Timeline\", \"Albums\", \"People\" and \"Settings\"." }));
  });
  
  await testmu.step("Assertion check", { instructionId: "1b194953-c9a5-4c47-bd29-141f6d1f221f" }, async () => {
    await testmu.verifyAssertion(page, "Assertion check", {"operator": ["equals"], "assertion_operands": [], "left_operand": null, "right_operand": null, "operands": [], "sub_results": [{"description": "the URL contains \"/timeline\".", "passed": true, "operator": "contains", "transforms": [], "json_path": null, "expected": "/timeline", "extracted_value": "{{timeline_url_check}}", "store_key": "timeline_url_check", "variable_refs": {"{{timeline_url_check}}": "http://localhost:8000/timeline"}}, {"description": "the sidebar shows \"Timeline\", \"Albums\", \"People\" and \"Settings\".", "passed": true, "operator": "equals", "transforms": [], "json_path": null, "expected": "true", "extracted_value": "{{sidebar_navigation_check}}", "store_key": "sidebar_navigation_check", "variable_refs": {"{{sidebar_navigation_check}}": "true"}}], "sub_checks": [{"description": "the URL contains \"/timeline\".", "store_key": "timeline_url_check", "expected_value": "/timeline", "extracted_value": "{{timeline_url_check}}", "operator": "contains", "transforms": []}, {"description": "the sidebar shows \"Timeline\", \"Albums\", \"People\" and \"Settings\".", "store_key": "sidebar_navigation_check", "expected_value": "true", "extracted_value": "{{sidebar_navigation_check}}", "operator": "equals", "transforms": []}], "composite_operator": "and", "claim": "the URL contains \"/timeline\" and the sidebar shows \"Timeline\", \"Albums\", \"People\" and \"Settings\"."});
  });
});

testmu.run();