from patchright.sync_api import sync_playwright, Cookie, Page
import json
import base64
import re
from typing import Optional, Dict, List, Any, TypedDict
from logging import getLogger, INFO
import time
import os

logger = getLogger(__name__)
logger.setLevel(INFO)

class ClientSession(TypedDict):
    auth_token: str | None
    bypass_akamai: Dict[str, Cookie]


class AuthenticationService:
    """
    Authentication Service class for e-banking application
    Handles user authentication, session management, and API interactions using Playwright
    """
    
    def __init__(self, base_url='https://online.simplii.com', cookies: List[Cookie] | None = None):
        """
        Initialize the authentication service
        
        Args:
            base_url: Base URL for the e-banking application
        """
        # Initialize Playwright
        self._playwright = None
        self._browser = None
        self._page: Page = None
        self._base_url = base_url
        
        # Authentication state
        self._is_authenticated = False
        
        # Storage key for session token
        self._session_token_key = "ebanking:session_token"
        
        # API endpoints
        self._sessions_api_endpoint = "/ebm-anp/api/v1/json/sessions"
        self._token_exchange_api_endpoint = "/ebm-anp/api/v1/json/getXAuthToken"
        self._validate_otvc_api_endpoint = "/ebm-anp/api/v1/json/validateOtvc"
        self._otvc_channel_api_endpoint = "/ebm-anp/api/v2/json/otvcchannel"
        
        # User and context information
        self._bank_brand = "pcf"
        self._customer_segment = ""
        self._customer_identifier = ""
        self._is_customer_data_input_required = None
        
        # Authentication token
        self._auth_token: str | None = None
        self._restored_cookies = cookies or []
    
    def initialize_browser(self):
        """Initialize the browser and context for making requests"""
        self._playwright = sync_playwright().start()
        self._browser = self._playwright.chromium.launch(headless=False)
        self._page = self._browser.new_page()
        
        # Set up storage state for session management
        self._page.context.add_cookies([{
            'name': self._session_token_key,
            'value': self._auth_token or '',
            'domain': self._extract_domain(self._base_url),
            'path': '/'
        }]) if self._auth_token else None

        self._page.context.add_cookies(self._restored_cookies)

        self._page.goto("https://www.simplii.com/")
        time.sleep(5)

        self._page.goto(self._base_url)
        time.sleep(20)

    def generate_session(self) -> ClientSession:
        if not self._page:
            raise ValueError("Browser not initialized. Call initialize_browser() first.")
        
        akamai_cookies = [
            cookie for cookie in self._page.context.cookies()
            if 'name' in cookie and cookie['name'] in ["bm_sv", "bm_sz", "_abck", "ak_bmsc", "AKA_A2", "bm_mi", "bmuid"]
        ]
        akamai_cookies = {cookie['name']: cookie for cookie in akamai_cookies}
        return {
            'bypass_akamai': akamai_cookies,
            'auth_token': self.authentication_token
        }
    
    
    def close_browser(self):
        """Close the browser and clean up resources"""
        if self._browser:
            self._browser.close()
        if self._playwright:
            self._playwright.stop()
        
        self._browser = None
        self._page = None
        self._playwright = None
    
    @staticmethod
    def _extract_domain(url):
        """Extract the domain from a URL"""
        match = re.search(r'https?://([^/]+)', url)
        return match.group(1) if match else 'localhost'
    
    @property
    def is_authenticated(self):
        """Check if the user is authenticated"""
        return self._is_authenticated
    
    @is_authenticated.setter
    def is_authenticated(self, value):
        """Set the authentication state"""
        self._is_authenticated = value
    
    @property
    def authentication_token(self):
        """Get the authentication token"""
        return self._auth_token
    
    @authentication_token.setter
    def authentication_token(self, token):
        """Set the authentication token"""
        self._auth_token = token.replace('"', '') if token else None
        
        # Update browser cookie if browser is initialized
        if self._page and self._auth_token:
            self._page.context.add_cookies([{
                'name': self._session_token_key,
                'value': self._auth_token,
                'domain': self._extract_domain(self._base_url),
                'path': '/'
            }])
    
    @property
    def bank_brand(self):
        """Get bank brand"""
        return self._bank_brand
    
    @bank_brand.setter
    def bank_brand(self, brand):
        """Set bank brand"""
        self._bank_brand = brand
    
    @property
    def customer_segment(self):
        """Get customer segment"""
        return self._customer_segment
    
    @customer_segment.setter
    def customer_segment(self, segment):
        """Set customer segment"""
        self._customer_segment = segment
    
    @property
    def customer_identifier(self):
        """Get customer identifier"""
        return self._customer_identifier
    
    @customer_identifier.setter
    def customer_identifier(self, identifier):
        """Set customer identifier"""
        self._customer_identifier = identifier
    
    @property
    def is_customer_data_input_required(self):
        """Check if customer data input is required"""
        return self._is_customer_data_input_required
    
    @is_customer_data_input_required.setter
    def is_customer_data_input_required(self, required):
        """Set if customer data input is required"""
        self._is_customer_data_input_required = required
    
    def _ensure_browser_initialized(self):
        """Ensure the browser is initialized before making requests"""
        if not self._browser:
            self.initialize_browser()
    
    def _make_request(self, method, url, headers=None, data=None):
        """
        Make an HTTP request using Playwright
        
        Args:
            method: HTTP method (GET, POST, DELETE)
            url: URL to request
            headers: Request headers
            data: Request data (for POST requests)
            
        Returns:
            Response object with data, headers, and status
        """
        self._ensure_browser_initialized()
        
        full_url = f"{self._base_url}{url}" if not url.startswith('http') else url
        
        # Prepare request options
        options = {
            'headers': headers or {},
            'timeout': 30000  # 30 seconds timeout
        }
        
        if data:
            options['data'] = json.dumps(data)
        
        # Make the request based on the method
        if method.lower() == 'get':
            response = self._page.context.request.get(full_url, **options)
        elif method.lower() == 'post':
            response = self._page.context.request.post(full_url, **options)
        elif method.lower() == 'delete':
            response = self._page.context.request.delete(full_url, **options)
        else:
            raise ValueError(f"Unsupported HTTP method: {method}")
        
        # Extract data from response
        response_data = {}
        try:
            response_data = response.json()
        except:
            # Not JSON response
            response_data = response.text()
            pass
        
        # Extract headers
        headers_dict = {}
        for header in response.headers.items():
            headers_dict[header[0].lower()] = header[1]
        
        return {
            'data': response_data,
            'headers': headers_dict,
            'status': response.status
        }
    
    def process_response_headers(self, headers):
        """
        Process response headers to extract authentication tokens and identifiers
        
        Args:
            headers: Response headers
        """
        # Update auth token if present in response
        if 'x-auth-token' in headers:
            self.authentication_token = headers['x-auth-token']
            self.is_authenticated = True
        
        # Update customer identifier if present
        if 'x-uci' in headers:
            self.customer_identifier = headers['x-uci']
    
    def set_client_data(self, response_data):
        """
        Update client data from API response
        
        Args:
            response_data: The data from API response
        """
        
        if 'segment' in response_data:
            self.customer_segment = response_data['segment']
        
        if 'cdiRequired' in response_data:
            self.is_customer_data_input_required = response_data['cdiRequired']
    
    def get_auth_token_from_z_token(self, z_token):
        """
        Exchange a Z-token for an authentication token
        
        Args:
            z_token: The Z-token to exchange
            
        Returns:
            Response from the token exchange API
        """
        headers = {
            'Z': z_token
            **self._build_auth_client_headers()
        }
        
        response = self._make_request(
            'get', 
            self._token_exchange_api_endpoint, 
            headers=headers
        )
        
        self.process_response_headers(response['headers'])
        return response
    
    def refresh_session(self, options=None):
        """
        Refresh the current session
        
        Args:
            options: Optional configuration for the refresh request
            
        Returns:
            Response from the session refresh API
            
        Raises:
            ValueError: If no authentication token is present
        """
        options = options or {}
        
        if not self.authentication_token:
            raise ValueError("Attempted to call refresh_session() with no authentication_token present")
        
        headers = {
            'Accept': 'application/json',
            'Content-Type': 'application/vnd.api+json',
            'X-Auth-Token': self.authentication_token,
            **self._build_auth_client_headers()
        }
        
        # Add any additional headers
        if 'headers' in options:
            headers.update(options['headers'])
        
        response = self._make_request(
            'get', 
            self._sessions_api_endpoint, 
            headers=headers
        )
        
        self.set_client_data(response['data'])
        self.process_response_headers(response['headers'])
        
        return response
    
    def _solve_2fa_challenge(self, transaction_id: str, delivery_channels: List[Dict[str, Any]]):
        preferred_channel = next(
            (channel for channel in delivery_channels if channel['preferred']), delivery_channels[0]
        )
        logger.info(f"Solving 2FA challenge using channel: {preferred_channel['deliveryChannel']}")

        headers = {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
            'X-Auth-Token': self.authentication_token,
            'Www-Authenticate': 'OTVC',
            **self._build_auth_client_headers()
        }
        
        data = {
            'deliveryChannel': {
                'channelValue': preferred_channel['channelValue'],
                'deliveryChannel': preferred_channel['deliveryChannel'],
                'preferred': preferred_channel['preferred']
            },
            'transactionId': transaction_id
        }

        logger.info(f"Requesting OTVC channel: {json.dumps(data)}")
        logger.info(f"Headers: {json.dumps(headers)}")

        # First make a request to the OTVC channel and set the webauth to OTVC
        response = self._make_request(
            "post",
            self._otvc_channel_api_endpoint,
            headers=headers,
            data=data
        )
        
        if response['status'] != 204:
            raise Exception(f"Failed to solve 2FA challenge: {response['data']}, status: {response['status']}, headers: {response['headers']}")
        
        # Read input from stdin for the OTP
        otp = input("Enter the OTP sent to your device: ")
        if not otp:
            raise ValueError("OTP cannot be empty")
        # Now make a request to validate the OTP
        response = self._make_request(
            "post",
            self._validate_otvc_api_endpoint,
            headers=headers,
            data={
                'channelType': preferred_channel['deliveryChannel'],
                'otvc': otp,
                'transactionId': transaction_id
            }
        )

        if response['status'] != 200:
            raise Exception(f"Failed to validate OTP: {response['data']}")

        return response

    def _build_auth_client_headers(self):
        """
        Build the client headers for the request
        
        Returns:
            Dictionary of headers
        """
        return {
            'Client-Type': 'DEFAULT_WEB',
            'brand': self._bank_brand
        }
    
    def authenticate_card_and_password(self, card_number, password, tracking_info=None, options=None):
        """
        Authenticate using card number and password
        
        Args:
            card_number: The card number
            password: The password
            tracking_info: Optional tracking information
            options: Additional request options
            
        Returns:
            Response from the authentication API
        """

        if self.authentication_token:
            # try to renew the session first
            try:
                return self.refresh_session()
            except Exception as e:
                logger.error(f"Failed to refresh session, attempting to re-authenticate")

        tracking_info = tracking_info or {}
        options = options or {}
        
        headers = {
            'Accept': 'application/vnd.api+json',
            'Content-Type': 'application/vnd.api+json',
            'Www-Authenticate': 'CardAndPassword',
            **self._build_auth_client_headers()
        }
        
        # Add any additional headers
        if 'headers' in options:
            headers.update(options['headers'])
        
        request_data = {
            'card': {
                'value': card_number,
                'encrypt': True,
                'encrypted': False
            },
            'password': password
        }
        
        # Add tracking IDs if provided
        if tracking_info.get('pageId'):
            request_data['pageId'] = tracking_info['pageId']
        
        if tracking_info.get('profilingId'):
            request_data['profilingId'] = tracking_info['profilingId']
        
        # Add any additional data
        if 'data' in options:
            request_data.update(options['data'])
        

        response = self._make_request(
            'post', 
            self._sessions_api_endpoint, 
            headers=headers, 
            data=request_data
        )
        if response['status'] == 401:
            response = self._handle_auth_401(response)
        elif response['status'] != 200:
            raise Exception(f"Authentication failed: {response['data']}, status: {response['status']}, headers: {response['headers']}")
        
        self.set_client_data(response['data'])
        self.process_response_headers(response['headers'])
        
        return response
    
    def _handle_auth_401(self, response):
        """
        Handle 401 Unauthorized response
        
        Args:
            response: The response object
        """
        # Even on 401 error, we might receive an auth token for further steps
        if 'x-auth-token' in response['headers']:
            self.authentication_token = response['headers']['x-auth-token']

        # Extract transaction ID and delivery channels from the response
        body = response['data']
        transaction_id = body.get('transactionId')
        delivery_channels = body.get('deliveryChannels', [])
        
        if not transaction_id or not delivery_channels:
            raise Exception("Invalid 401 response: Missing transaction ID or delivery channels")
        
        # Solve the 2FA challenge
        return self._solve_2fa_challenge(transaction_id, delivery_channels)

# Example usage
def main():
    # Create an instance of the authentication service
    session_file = 'simplii_session.json'
    cookies: List[Cookie] = []
    auth_token: str | None = None
    # Load cookies from a file or other source if needed
    if os.path.exists(session_file):
        with open(session_file, 'r') as f:
            session_data: ClientSession = json.load(f)
            cookies = [v for _, v in session_data['bypass_akamai'].items()]
            auth_token = session_data['auth_token']
    auth_service = AuthenticationService(cookies=cookies)
    auth_service.authentication_token = auth_token
    
    try:
        # Login with card and password
        _ = auth_service.authenticate_card_and_password(
            card_number='|FILL_CARD_HERE|',
            password='|FILL_PASSWORD_HERE|',
            tracking_info={}
        )
        
        print(f"Authentication successful: {auth_service.is_authenticated}")
        print(f"Auth token: {auth_service.authentication_token}")
        
        # Refresh the session
        refresh_response = auth_service.refresh_session()
        print(f"Session refreshed: {refresh_response['status'] == 200}")

        with open(session_file, 'w') as f:
            json.dump(auth_service.generate_session(), f, indent=4)
        
    finally:
        # Clean up resources
        auth_service.close_browser()


# Run the example
if __name__ == "__main__":
    main()