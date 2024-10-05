<template>
    <!-- <h1 class="title">Medium section</h1> -->
    <div class="mailchimp">
        <!-- Begin Mailchimp Signup Form -->
        <link href="//cdn-images.mailchimp.com/embedcode/horizontal-slim-10_7.css" rel="stylesheet" type="text/css">
        <div id="mc_embed_signup">
            <form action="https://gmail.us18.list-manage.com/subscribe/post?u=2c6b66f6b94ee6dfbf9524986&amp;id=36cf9e0d6c"
                method="post" id="mc-embedded-subscribe-form" name="mc-embedded-subscribe-form" class="validate"
                target="_blank" novalidate>
                <div id="mc_embed_signup_scroll">
                    <label for="mce-EMAIL">
                        <p class="subtitle has-text-centered">Sign up now and get early access when we launch.</p>
                    </label>
                    <br>
                    <b-input type="email" value="" name="EMAIL" v-model="email" class="email" id="mce-EMAIL"
                        placeholder="email address" required />
                    <!-- real people should not fill this in and expect good things - do not remove this or risk form bot signups-->
                    <div style="position: absolute; left: -5000px;" aria-hidden="true">
                        <input type="text" name="b_2c6b66f6b94ee6dfbf9524986_36cf9e0d6c" tabindex="-1" value="">
                    </div>
                    <div align="center">
                        <input type="submit" value="Subscribe" name="subscribe" id="mc-embedded-subscribe"
                            class="button subscribe-button is-success is-medium is-block">
                    </div>
                </div>
            </form>
        </div>
        <!--End mc_embed_signup-->
    </div>
</template>
<script type="text/javascript">
export default {
    name: 'mailchimp',
    data() {
        return {
            email: ""
        };
    },
}
</script>
<style>
.mailchimp {
    width: 400px;
}

#mc_embed_signup {
    clear: left;
    font: 14px Inter, Arial, sans-serif;
    width: 100%;
}
</style>