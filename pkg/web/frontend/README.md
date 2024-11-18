# TAF Kafka Watch Frontend

## 💡 Usage

### 🚀 Starting the Development Server

To start the development server with hot-reload, run the following command. The server will be accessible at [http://localhost:3001](http://localhost:3001):

```bash
npm run dev
```

It will additionally proxy WebSocket connections to the go portion of the web interface, assuming to be listening on port 3000 ([http://localhost:3000](http://localhost:3000)).

### 🛠️ Building for Production

To build your frontend to statically embed inside the code binary:

```bash
npm run build
```

The `dist` directory contains the compiled frontend, which will also be embedded in the go binary.
There for the contents of the `dist` directory are tracked in GIT, so that go application can be compiled, even if someone freshly checked out the repository.
