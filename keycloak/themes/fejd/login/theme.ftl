<#import "template.ftl" as layout>
<@layout.registrationLayout displayInfo=false displayMessage=false; section>
  <#if section = "header">
    <h1>Welcome to FEJD</h1>
  <#elseif section = "form">
    <form class="login-form" action="${url.loginAction}" method="post">
      <input type="hidden" name="credentialId" value="${credentialId!}" />
      <div>
        <label for="username">Username</label>
        <input id="username" name="username" value="${(login.username!'')}" autocomplete="username" />
      </div>
      <div>
        <label for="password">Password</label>
        <input id="password" type="password" name="password" autocomplete="current-password" />
      </div>
      <#if message?has_content>
        <p class="error">${message.summary}</p>
      </#if>
      <button type="submit">Sign in</button>
    </form>
  </#if>
</@layout.registrationLayout>
